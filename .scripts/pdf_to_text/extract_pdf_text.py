import sys
import argparse
import re
import unicodedata
from pathlib import Path
from typing import List, Tuple, Dict

try:
    from pdfminer.high_level import extract_text
except Exception as e:
    print(f"Failed to import pdfminer: {e}", file=sys.stderr)
    sys.exit(2)


def read_pdf_text(path: Path) -> str:
    try:
        return extract_text(str(path)) or ""
    except Exception as e:
        return f"[ERROR] Could not extract text from {path.name}: {e}\n"


def collect_pdfs(directory: Path) -> List[Path]:
    return sorted([p for p in directory.glob("*.pdf") if p.is_file()])


def _strip_accents(text: str) -> str:
    return "".join(
        ch for ch in unicodedata.normalize("NFKD", text) if not unicodedata.combining(ch)
    )


def slugify(text: str) -> str:
    s = _strip_accents(text).lower()
    s = s.replace("&", " and ")
    s = re.sub(r"[^a-z0-9]+", "-", s)
    s = re.sub(r"-+", "-", s).strip("-")
    return s or "section"


def normalize_text(text: str) -> str:
    # Trim trailing spaces and collapse >2 blank lines to 2 for readability
    lines = [ln.rstrip() for ln in text.splitlines()]
    out_lines: List[str] = []
    blank_streak = 0
    for ln in lines:
        if ln.strip() == "":
            blank_streak += 1
            if blank_streak <= 2:
                out_lines.append("")
        else:
            blank_streak = 0
            out_lines.append(ln)
    return "\n".join(out_lines).strip() + "\n"


def ensure_dir(path: Path) -> None:
    path.mkdir(parents=True, exist_ok=True)


def build_output_path(pdf: Path, out_dir: Path, fmt: str) -> Path:
    ext = ".md" if fmt == "md" else ".txt"
    return out_dir / (pdf.stem + ext)


def render_markdown(text: str, pdf_name: str) -> str:
    def is_numeric_heading(s: str) -> Tuple[int, str] | None:
        m = re.match(r"^(\d+(?:\.\d+){0,3})\s+(.+)$", s)
        if not m:
            return None
        parts = m.group(1).split(".")
        # Level 2 for X, 3 for X.Y, 4 for X.Y.Z, capped at 6
        level = min(6, 1 + len(parts))
        return level, s

    def is_bullet(s: str) -> bool:
        return s.startswith("•") or s.startswith("o ") or s.startswith("▪") or s.startswith("- ")

    TOC_NAMES = {"table of contents", "sisällys", "sisallys"}
    UNNUMBERED_H2 = {
        "objective",
        "purpose & scope",
        "workflows",
        "users, roles, and ui permissions",
        "core experiences",
        "navigation, routing, and flow",
        "data fetching & state",
        "ux quality bars",
        "security & compliance in the ui",
        "telemetry & diagnostics (front-end events)",
        "acceptance criteria (sampling)",
        "open questions",
        "editor api",
        "document list api",
        "document preview api",
        "capabilities & attachments",
        "capabilities",
        "attachments",
        "tags",
        "rules & validation (server-enforced)",
        "auth roles",
        "errors",
    }

    def parse_headings(lines: List[str]) -> List[Tuple[int, int, str]]:
        result: List[Tuple[int, int, str]] = []
        for idx, raw in enumerate(lines):
            s = raw.strip()
            if not s or is_bullet(s):
                continue
            low = s.lower()
            if low in TOC_NAMES:
                result.append((idx, 2, s))
                continue
            num = is_numeric_heading(s)
            if num:
                level, text_full = num
                result.append((idx, level, text_full))
                continue
            if low in UNNUMBERED_H2:
                result.append((idx, 2, s))
        return result

    def build_anchor_map(headings: List[Tuple[int, int, str]]) -> Dict[int, str]:
        used: Dict[str, int] = {}
        anchors: Dict[int, str] = {}
        for idx, _level, text in headings:
            base = slugify(text)
            count = used.get(base, 0)
            anchor = f"{base}-{count+1}" if count > 0 else base
            used[base] = count + 1
            anchors[idx] = anchor
        return anchors

    def is_toc_entry_line(s: str) -> bool:
        # lines like: Title .......... 2
        return bool(re.search(r"\.\.{2,}\s*\d+\s*$", s))

    def build_toc_lines(headings: List[Tuple[int, int, str]], anchor_map: Dict[int, str]) -> List[str]:
        lines: List[str] = []
        for idx, level, text_h in headings:
            low = text_h.strip().lower()
            if low in TOC_NAMES:
                continue
            if level < 2 or level > 6:
                continue
            indent = "  " * (level - 2)
            anchor = anchor_map.get(idx, slugify(text_h))
            lines.append(f"{indent}- [{text_h}](#{anchor})")
        lines.append("")
        return lines

    def transform(text_in: str) -> str:
        body_lines = text_in.splitlines()

        # Locate Table of Contents block to exclude it from heading discovery
        n = len(body_lines)
        toc_start = -1
        toc_end = -1
        for idx, raw in enumerate(body_lines):
            if raw.strip().lower() in TOC_NAMES:
                toc_start = idx
                j = idx + 1
                while j < n and (body_lines[j].strip() == "" or is_toc_entry_line(body_lines[j].strip())):
                    j += 1
                toc_end = j  # exclusive
                break

        if toc_start >= 0:
            scan_lines = body_lines[:toc_start] + body_lines[toc_end:]
        else:
            scan_lines = body_lines

        headings = parse_headings(scan_lines)

        # Map heading positions back to original indices when ToC was removed
        if toc_start >= 0:
            adjusted: List[Tuple[int, int, str]] = []
            for idx, lvl, txt in headings:
                true_idx = idx if idx < toc_start else idx + (toc_end - toc_start)
                adjusted.append((true_idx, lvl, txt))
            headings = adjusted

        anchor_map = build_anchor_map(headings)
        heading_lookup: Dict[int, Tuple[int, str]] = {i: (lvl, txt) for i, lvl, txt in headings}

        out: List[str] = []
        i = 0
        while i < n:
            s_raw = body_lines[i]
            s = s_raw.strip()
            low = s.lower()
            if low in TOC_NAMES:
                # Insert ToC heading with anchor and generated list
                anchor = anchor_map.get(i, slugify(s))
                out.append(f"<a id=\"{anchor}\"></a>")
                out.append(f"## {s}")
                # skip original toc entries
                j = i + 1
                while j < n and (body_lines[j].strip() == "" or is_toc_entry_line(body_lines[j].strip())):
                    j += 1
                out.extend(build_toc_lines(headings, anchor_map))
                i = j
                continue

            if i in heading_lookup:
                level, text_h = heading_lookup[i]
                anchor = anchor_map.get(i, slugify(text_h))
                level = max(2, min(6, level))
                out.append(f"<a id=\"{anchor}\"></a>")
                out.append(f"{'#' * level} {text_h}")
                i += 1
                continue

            out.append(s_raw)
            i += 1

        return "\n".join(out).strip() + "\n"

    header = [
        f"# {pdf_name} (extracted)",
        "",
        "> Extracted via pdfminer.six; formatting may differ from the original.",
        "",
        "---",
        "",
    ]
    transformed = transform(text)
    return "\n".join(header) + transformed


def save_extracted(text: str, pdf: Path, out_dir: Path, fmt: str) -> Path:
    cleaned = normalize_text(text)
    if fmt == "md":
        contents = render_markdown(cleaned, pdf.name)
    else:
        contents = cleaned
    out_path = build_output_path(pdf, out_dir, fmt)
    out_path.write_text(contents, encoding="utf-8", errors="replace")
    return out_path


def main() -> int:
    # Ensure we can print unicode safely on Windows consoles
    try:
        sys.stdout.reconfigure(encoding="utf-8", errors="replace")
    except Exception:
        pass
    parser = argparse.ArgumentParser(description="Extract text from PDFs in a directory")
    parser.add_argument("directory", nargs="?", default=".", help="Directory to scan for PDFs")
    parser.add_argument("--max-chars", type=int, default=0, help="Max chars to print per file (0 = suppress)")
    parser.add_argument("--name", action="append", help="Only process PDFs matching this name (exact match)")
    parser.add_argument("--out-dir", default=None, help="Directory to write extracted files (default: <dir>/extracted_text)")
    parser.add_argument("--format", choices=["md", "txt"], default="md", help="Output format for saved files")
    parser.add_argument("--combined", action="store_true", help="Also write a combined file of all PDFs")
    args = parser.parse_args()

    dir_path = Path(args.directory).resolve()
    if not dir_path.exists() or not dir_path.is_dir():
        print(f"Directory not found: {dir_path}", file=sys.stderr)
        return 1

    pdfs = collect_pdfs(dir_path)
    if args.name:
        names = set(args.name)
        pdfs = [p for p in pdfs if p.name in names]
    if not pdfs:
        print("No PDF files found.")
        return 0

    # Prepare output directory
    out_dir = Path(args.out_dir) if args.out_dir else (dir_path / "extracted_text")
    out_dir = out_dir.resolve()
    ensure_dir(out_dir)

    combined_parts: List[str] = []
    combined_sep = "\n\n" + ("#" * 80) + "\n\n"

    for pdf in pdfs:
        text = read_pdf_text(pdf)
        saved_path = save_extracted(text, pdf, out_dir, args.format)
        preview = ""
        if args.max_chars > 0:
            preview_text = text[: args.max_chars]
            preview = f"\nPreview (first {args.max_chars} chars):\n" + preview_text + ("\n[...truncated...]" if len(text) > args.max_chars else "")
        print(f"Saved: {pdf.name} -> {saved_path} (chars: {len(text)})" + preview)

        if args.combined:
            cleaned = normalize_text(text)
            if args.format == "md":
                combined_parts.append(render_markdown(cleaned, pdf.name))
            else:
                combined_parts.append(f"FILE: {pdf.name}\n\n" + cleaned)

    if args.combined:
        combined_ext = ".md" if args.format == "md" else ".txt"
        combined_path = out_dir / ("combined_extracted" + combined_ext)
        combined_content = combined_sep.join(combined_parts)
        combined_path.write_text(combined_content, encoding="utf-8", errors="replace")
        print(f"Combined file written: {combined_path}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
