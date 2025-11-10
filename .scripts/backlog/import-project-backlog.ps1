Param(
  [string]$Owner = "Jannnesi",
  [int]$ProjectNumber = 3,
  [string]$RepoOwner = "Jannnesi",
  [string]$RepoName = "Palvelinohjelmointi-GO-Group-2",
  [string]$JsonPath = "project-items.json",
  [switch]$Draft,
  [switch]$DryRun,
  [int]$MaxItems = 0,
  [int]$Skip = 0,
  [switch]$LinkRepo
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

function Write-Info($msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Write-Do($msg) { Write-Host "[DO]   $msg" -ForegroundColor Yellow }
function Write-Success($msg) { Write-Host "[OK]   $msg" -ForegroundColor Green }
function Write-Warn($msg) { Write-Host "[WARN] $msg" -ForegroundColor DarkYellow }
function Write-Err($msg) { Write-Host "[ERR]  $msg" -ForegroundColor Red }

function Ensure-Gh() {
  if (-not (Get-Command gh -ErrorAction SilentlyContinue)) {
    throw "GitHub CLI 'gh' not found in PATH. Install via winget/choco/scoop."
  }
  try {
    $status = & gh auth status 2>&1 | Out-String
    if ($status -match 'Logged in to') {
      Write-Info ($status.Trim())
    } else {
      Write-Warn "gh not authenticated. Run: gh auth login"
    }
  } catch {
    Write-Warn "Could not verify gh auth status: $_"
  }
}

function Get-ProjectAndFields([string]$owner, [int]$number) {
  Write-Info "Fetching project $number for owner '$owner'"
  $projJson = & gh project view $number --owner $owner --format json | Out-String
  $proj = $projJson | ConvertFrom-Json
  if (-not $proj) { throw "Project not found or no access." }

  $fieldsJson = & gh project field-list $number --owner $owner --format json | Out-String
  $fields = ($fieldsJson | ConvertFrom-Json).fields
  if (-not $fields) { throw "Could not read project fields." }

  $statusField = $fields | Where-Object { $_.name -eq 'Status' }
  if (-not $statusField) { throw "Status field not found in project." }
  $statusBacklog = $statusField.options | Where-Object { $_.name -eq 'Backlog' }
  if (-not $statusBacklog) { throw "Backlog option not found in Status field." }
  $estimateField = $fields | Where-Object { $_.name -eq 'Estimate' }
  if (-not $estimateField) { Write-Warn "Estimate field not found; estimates will be skipped." }

  return [pscustomobject]@{
    ProjectId = $proj.id
    ProjectNumber = $proj.number
    Title = $proj.title
    Fields = $fields
    StatusFieldId = $statusField.id
    StatusBacklogOptionId = $statusBacklog.id
    EstimateFieldId = $estimateField.id
  }
}

function Ensure-RepoLinked([int]$projectNumber, [string]$owner, [string]$repoName) {
  if ($DryRun) { Write-Do "gh project link $projectNumber --owner $owner --repo $repoName"; return }
  try {
    & gh project link $projectNumber --owner $owner --repo $repoName | Out-Null
    Write-Success "Linked repo '$repoName' to project $projectNumber ($owner)"
  } catch {
    Write-Warn "Could not link repo: $_"
  }
}

function Ensure-LabelsExist([string]$repo, [string[]]$labels) {
  if (-not $labels -or $labels.Count -eq 0) { return }
  $unique = $labels | Where-Object { $_ -and $_.Trim() -ne '' } | ForEach-Object { $_.Trim() } | Select-Object -Unique
  if ($unique.Count -eq 0) { return }

  $existing = @()
  try {
    $list = & gh label list -R $repo --limit 200 --json name | ConvertFrom-Json
    if ($list) { $existing = $list | Select-Object -ExpandProperty name }
  } catch {
    Write-Warn "Could not list existing labels for ${repo}: $_"
  }

  $missing = @()
  foreach ($name in $unique) {
    if ($existing -notcontains $name) { $missing += $name }
  }

  foreach ($name in $missing) {
    $args = @('label','create',$name,'-R',$repo,'--color','1F77B4','--description',"Auto-created by backlog importer")
    if ($DryRun) {
      Write-Do ("gh " + ($args -join ' '))
    } else {
      $out = & gh @args 2>&1
      $exit = $LASTEXITCODE
      if ($exit -ne 0) {
        $msg = ($out | Out-String).Trim()
        Write-Warn "Could not create label '$name': $msg"
      } else {
        Write-Success "Created label '$name'"
      }
    }
  }
}

function Build-Body($item) {
  $lines = @()
  if ($item.description) { $lines += "$($item.description)" }
  if ($item.acceptance_criteria) {
    $lines += ''
    $lines += 'Acceptance Criteria:'
    foreach ($ac in $item.acceptance_criteria) { $lines += "- $ac" }
  }
  if ($item.depends_on -and $item.depends_on.Count -gt 0) {
    $lines += ''
    $lines += ("Depends on: " + ($item.depends_on -join ', '))
  }
  if ($item.id) {
    $lines += ''
    $lines += "Internal ID: $($item.id)"
  }
  return ($lines -join "`n")
}

function Create-IssueAndAddToProject($item, $repo, $projectNumber, $projectId, $owner, $statusFieldId, $statusBacklogId, $estimateFieldId) {
  $body = Build-Body $item
  $labelArgs = @()
  if ($item.labels) {
    foreach ($lab in $item.labels) { $labelArgs += @('--label', [string]$lab) }
  }

  $issueUrl = $null
  $issueNum = $null

  $createArgs = @('issue','create','-R',$repo,'--title',[string]$item.title,'--body',[string]$body) + $labelArgs
  if ($DryRun) {
    Write-Do ("gh " + ($createArgs -join ' '))
    $issueUrl = "https://github.com/$repo/issues/<NEW>"
    $issueNum = -1
  } else {
    $out = & gh @createArgs 2>&1
    $exit = $LASTEXITCODE
    $outStr = ($out | Out-String).Trim()
    if ($exit -ne 0) {
      throw "gh issue create failed: $outStr"
    }
    $issueUrl = ($outStr -split "`n" | Select-Object -Last 1).Trim()
    if (-not $issueUrl) { throw "Failed to parse issue URL from gh output: $outStr" }
    $issueNum = [int]($issueUrl.Split('/') | Select-Object -Last 1)
    Write-Success "Created issue #$issueNum"
  }

  $addArgs = @('project','item-add',$projectNumber,'--owner',$owner,'--url',$issueUrl,'--format','json')
  $projectItemId = $null
  if ($DryRun) {
    Write-Do ("gh " + ($addArgs -join ' '))
    $projectItemId = '<PROJECT_ITEM_ID>'
  } else {
    $add = (& gh @addArgs | ConvertFrom-Json)
    $projectItemId = $add.id
    if (-not $projectItemId) { throw "Failed to read project item id for issue #$issueNum" }
    Write-Success "Added issue #$issueNum to project (item $projectItemId)"
  }

  # Set Status = Backlog
  $editStatusArgs = @('project','item-edit','--id',$projectItemId,'--project-id',$projectId,'--field-id',$statusFieldId,'--single-select-option-id',$statusBacklogId)
  if ($DryRun) {
    Write-Do ("gh " + ($editStatusArgs -join ' '))
  } else {
    & gh @editStatusArgs | Out-Null
  }

  # Set Estimate if available
  if ($estimateFieldId -and $item.estimate -ne $null -and $item.estimate -ne '') {
    $num = [double]$item.estimate
    $editEstArgs = @('project','item-edit','--id',$projectItemId,'--project-id',$projectId,'--field-id',$estimateFieldId,'--number',$num)
    if ($DryRun) {
      Write-Do ("gh " + ($editEstArgs -join ' '))
    } else {
      & gh @editEstArgs | Out-Null
    }
  }

  return [pscustomobject]@{ IssueNumber = $issueNum; IssueUrl = $issueUrl; ProjectItemId = $projectItemId }
}

function Create-DraftItem($item, $projectNumber, $projectId, $owner, $statusFieldId, $statusBacklogId, $estimateFieldId) {
  $body = Build-Body $item
  $createArgs = @('project','item-create',$projectNumber,'--owner',$owner,'--title',[string]$item.title,'--body',[string]$body,'--format','json')
  $projectItemId = $null
  if ($DryRun) {
    Write-Do ("gh " + ($createArgs -join ' '))
    $projectItemId = '<DRAFT_ITEM_ID>'
  } else {
    $draft = (& gh @createArgs | ConvertFrom-Json)
    $projectItemId = $draft.id
    if (-not $projectItemId) { throw "Failed to create draft item for '$($item.title)'" }
    Write-Success "Created draft item $projectItemId"
  }

  # Set Status = Backlog
  $editStatusArgs = @('project','item-edit','--id',$projectItemId,'--project-id',$projectId,'--field-id',$statusFieldId,'--single-select-option-id',$statusBacklogId)
  if ($DryRun) {
    Write-Do ("gh " + ($editStatusArgs -join ' '))
  } else {
    & gh @editStatusArgs | Out-Null
  }

  # Set Estimate
  if ($estimateFieldId -and $item.estimate -ne $null -and $item.estimate -ne '') {
    $num = [double]$item.estimate
    $editEstArgs = @('project','item-edit','--id',$projectItemId,'--project-id',$projectId,'--field-id',$estimateFieldId,'--number',$num)
    if ($DryRun) {
      Write-Do ("gh " + ($editEstArgs -join ' '))
    } else {
      & gh @editEstArgs | Out-Null
    }
  }

  return [pscustomobject]@{ IssueNumber = $null; IssueUrl = $null; ProjectItemId = $projectItemId }
}

function Add-DependencyComments($items, $issueMap, $repo) {
  foreach ($item in $items) {
    if (-not $item.depends_on -or $item.depends_on.Count -eq 0) { continue }
    if (-not $issueMap.ContainsKey($item.id)) { continue }
    $thisIssue = $issueMap[$item.id]
    if (-not $thisIssue.IssueNumber -or $thisIssue.IssueNumber -lt 0) { continue }

    $depNums = @()
    foreach ($depId in $item.depends_on) {
      if ($issueMap.ContainsKey($depId) -and $issueMap[$depId].IssueNumber -and $issueMap[$depId].IssueNumber -gt 0) {
        $depNums += "#" + $issueMap[$depId].IssueNumber
      }
    }
    if ($depNums.Count -eq 0) { continue }

    $comment = "Depends on " + ($depNums -join ' ')
    $args = @('issue','comment','-R',$repo,[string]$thisIssue.IssueNumber,'--body',[string]$comment)
    if ($DryRun) {
      Write-Do ("gh " + ($args -join ' '))
    } else {
      & gh @args | Out-Null
      Write-Success "Commented dependencies on #$($thisIssue.IssueNumber): $comment"
    }
  }
}

try {
  Ensure-Gh

  $repo = "$RepoOwner/$RepoName"

  $meta = Get-ProjectAndFields -owner $Owner -number $ProjectNumber
  Write-Info "Project: #$($meta.ProjectNumber) '$($meta.Title)' (id $($meta.ProjectId))"
  Write-Info "Status field id: $($meta.StatusFieldId); Backlog option id: $($meta.StatusBacklogOptionId)"
  if ($meta.EstimateFieldId) { Write-Info "Estimate field id: $($meta.EstimateFieldId)" }

  if ($LinkRepo) { Ensure-RepoLinked -projectNumber $ProjectNumber -owner $Owner -repoName $RepoName }

  if (-not (Test-Path -LiteralPath $JsonPath)) { throw "JSON file not found: $JsonPath" }
  $data = Get-Content -Raw -LiteralPath $JsonPath | ConvertFrom-Json
  if (-not $data.backlog) { throw "No 'backlog' array found in $JsonPath" }
  $items = @($data.backlog)
  if ($Skip -gt 0) { $items = $items | Select-Object -Skip $Skip }
  if ($MaxItems -gt 0) { $items = $items | Select-Object -First $MaxItems }

  Write-Info ("Items to process: " + $items.Count)

  # Ensure required labels exist when creating issues
  if (-not $Draft) {
    $allLabels = @()
    foreach ($it in $items) { if ($it.labels) { $allLabels += $it.labels } }
    Ensure-LabelsExist -repo $repo -labels $allLabels
  }

  $issueMap = @{}

  # Pass 1: create issues/drafts, add to project, set fields
  foreach ($it in $items) {
    Write-Info "Processing: $($it.id) - $($it.title)"
    if ($Draft) {
      $res = Create-DraftItem -item $it -projectNumber $ProjectNumber -projectId $meta.ProjectId -owner $Owner -statusFieldId $meta.StatusFieldId -statusBacklogId $meta.StatusBacklogOptionId -estimateFieldId $meta.EstimateFieldId
    } else {
      $res = Create-IssueAndAddToProject -item $it -repo $repo -projectNumber $ProjectNumber -projectId $meta.ProjectId -owner $Owner -statusFieldId $meta.StatusFieldId -statusBacklogId $meta.StatusBacklogOptionId -estimateFieldId $meta.EstimateFieldId
    }
    $issueMap[$it.id] = $res
  }

  # Pass 2: dependency comments (only when issues are created)
  if (-not $Draft) {
    Add-DependencyComments -items $items -issueMap $issueMap -repo $repo
  }

  Write-Success "Completed. Processed $($items.Count) item(s)."
} catch {
  Write-Err $_
  exit 1
}
