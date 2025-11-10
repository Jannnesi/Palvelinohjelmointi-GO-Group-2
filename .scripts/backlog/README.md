# Project Backlog Importer (GitHub Projects v2)

PowerShell script to import backlog items from a JSON file into a GitHub Project (Projects v2). Supports two modes:
- Create GitHub Issues in a repository and add them to the project (recommended if you want labels and dependency comments)
- Create Project Draft Items (no repository issues)

## Requirements
- GitHub CLI installed: `gh --version`
- Authenticated with `project` scope:
  - Check: `gh auth status`
  - Ensure scope: `gh auth refresh -s project`
- Access to the target Project (owner + project number)
- Optional: write access to target repository when creating issues

## Script
- Path: `scripts/import-project-backlog.ps1`
- Input JSON: `project-items.json` (default path is repo root)

The JSON is expected to have a `backlog` array of items with fields: `id`, `title`, `status`, `estimate`, `labels`, `description`, `acceptance_criteria`, `depends_on`.

## Quick Start (Issues mode)
This creates issues in the repo, adds them to the project, sets Status=Backlog and Estimate, and comments dependencies. Missing labels referenced by items are auto-created in the repo.

```
pwsh -File scripts/import-project-backlog.ps1 \
  -Owner "Jannnesi" \
  -ProjectNumber 3 \
  -RepoOwner "Jannnesi" \
  -RepoName "Palvelinohjelmointi-GO-Group-2" \
  -JsonPath "project-items.json" \
  -LinkRepo
```

Preview (no changes):
```
pwsh -File scripts/import-project-backlog.ps1 -Owner "Jannnesi" -ProjectNumber 3 -RepoOwner "Jannnesi" -RepoName "Palvelinohjelmointi-GO-Group-2" -JsonPath "project-items.json" -DryRun -MaxItems 1
```

## Quick Start (Drafts mode)
Creates Project Draft Items only (no repository issues, labels won’t apply, and dependency comments are skipped):
```
pwsh -File scripts/import-project-backlog.ps1 -Owner "Jannnesi" -ProjectNumber 3 -JsonPath "project-items.json" -Draft
```

## Parameters
- `-Owner` (string): Project owner login (e.g., `Jannnesi`).
- `-ProjectNumber` (int): Project number from the URL (e.g., `.../projects/3`).
- `-RepoOwner` (string): Repository owner (issues mode only).
- `-RepoName` (string): Repository name (issues mode only).
- `-JsonPath` (string): Path to JSON input (default `project-items.json`).
- `-Draft` (switch): Create draft items instead of issues.
- `-DryRun` (switch): Print intended `gh` commands without making changes.
- `-MaxItems` (int): Only process the first N items (useful for testing).
- `-Skip` (int): Skip the first N items (useful to resume after a partial run).
- `-LinkRepo` (switch): Link the repository to the project before import.

## What the script does
- Fetches project and field IDs (Status and Estimate) automatically.
- Issues mode:
  - Creates issues with labels and a body composed from description, acceptance criteria, depends_on, and internal ID.
  - Adds issues to the project and sets Status=Backlog and Estimate.
  - Adds a comment like `Depends on #<num> #<num>` when inter-item dependencies exist.
  - Auto-creates any missing labels in the repo (default color `1F77B4`). Pre-create labels yourself to control colors.
- Drafts mode:
  - Creates draft items, sets Status=Backlog and Estimate. Labels and dependency comments do not apply.

## Notes and Safety
- Not idempotent: re-running will create duplicate issues/items. To resume a partial run, combine `-Skip` and `-MaxItems`.
- Consider using `-DryRun` first to validate commands and field mapping.
- Labels are only visible when the project item is a repository Issue.

## Validation
- List project items:
  - `gh project item-list 3 --owner "Jannnesi"`
- Inspect repository issues:
  - `gh issue list -R "Jannnesi/Palvelinohjelmointi-GO-Group-2"`

## Troubleshooting
- Missing `gh`: install via `winget install GitHub.cli`, `choco install gh -y`, or `scoop install gh`.
- Insufficient scope: `gh auth refresh -s project`.
- Field errors (e.g., Status/Estimate not found): ensure your project has a single-select Status with an option named `Backlog`, and a number field named `Estimate`.
