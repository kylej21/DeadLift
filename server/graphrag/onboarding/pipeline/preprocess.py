import logging
from pathlib import Path

log = logging.getLogger(__name__)

INCLUDE_EXTENSIONS = {
    ".py", ".js", ".ts", ".go", ".java", ".rb", ".rs", ".cs",
    ".cpp", ".c", ".h", ".yaml", ".yml", ".json", ".toml", ".md", ".sh",
}

SKIP_DIRS = {
    ".git", "node_modules", "__pycache__", ".venv", "vendor", "dist", "build",
}


def _sanitize_name(relative_path: str) -> str:
    return relative_path.replace("/", "__").replace("\\", "__") + ".txt"


def _write_file(source: Path, relative_path: str, input_dir: Path):
    out_path = input_dir / _sanitize_name(relative_path)
    content = f"# File: {relative_path}\n" + source.read_text(errors="replace")
    out_path.write_text(content, encoding="utf-8")


def preprocess_all(repo_path: str, input_dir: str):
    repo = Path(repo_path)
    out = Path(input_dir)
    out.mkdir(parents=True, exist_ok=True)
    log.info("Preprocessing all files: repo=%s output=%s", repo, out)

    count = skipped = 0
    for file in repo.rglob("*"):
        if not file.is_file():
            continue
        parts = set(file.relative_to(repo).parts)
        if parts & SKIP_DIRS:
            log.debug("Skipping (excluded dir): %s", file)
            skipped += 1
            continue
        if file.suffix not in INCLUDE_EXTENSIONS:
            log.debug("Skipping (extension %s): %s", file.suffix, file)
            skipped += 1
            continue
        rel = file.relative_to(repo).as_posix()
        log.debug("Writing: %s", rel)
        _write_file(file, rel, out)
        count += 1

    log.info("Preprocessing complete: %d written, %d skipped", count, skipped)


def preprocess_files(repo_path: str, input_dir: str, files: list[str]):
    repo = Path(repo_path)
    out = Path(input_dir)
    out.mkdir(parents=True, exist_ok=True)
    log.info("Preprocessing %d changed file(s): %s", len(files), files)

    count = skipped = 0
    for rel in files:
        file = repo / rel
        if not file.exists() or not file.is_file():
            log.debug("Skipping (not found): %s", rel)
            skipped += 1
            continue
        parts = set(Path(rel).parts)
        if parts & SKIP_DIRS:
            log.debug("Skipping (excluded dir): %s", rel)
            skipped += 1
            continue
        if file.suffix not in INCLUDE_EXTENSIONS:
            log.debug("Skipping (extension %s): %s", file.suffix, rel)
            skipped += 1
            continue
        log.debug("Writing: %s", rel)
        _write_file(file, rel, out)
        count += 1

    log.info("Preprocessing complete: %d written, %d skipped", count, skipped)
