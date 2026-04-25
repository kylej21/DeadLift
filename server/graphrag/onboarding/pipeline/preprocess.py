from pathlib import Path

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

    for file in repo.rglob("*"):
        if not file.is_file():
            continue
        parts = set(file.relative_to(repo).parts)
        if parts & SKIP_DIRS:
            continue
        if file.suffix not in INCLUDE_EXTENSIONS:
            continue
        rel = file.relative_to(repo).as_posix()
        _write_file(file, rel, out)


def preprocess_files(repo_path: str, input_dir: str, files: list[str]):
    repo = Path(repo_path)
    out = Path(input_dir)
    out.mkdir(parents=True, exist_ok=True)

    for rel in files:
        file = repo / rel
        if not file.exists() or not file.is_file():
            continue
        parts = set(Path(rel).parts)
        if parts & SKIP_DIRS:
            continue
        if file.suffix not in INCLUDE_EXTENSIONS:
            continue
        _write_file(file, rel, out)
