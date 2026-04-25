import subprocess
from pathlib import Path


def clone_repo(repo_url: str, dest_dir: str) -> str:
    dest = Path(dest_dir)
    dest.parent.mkdir(parents=True, exist_ok=True)
    subprocess.run(["git", "clone", repo_url, str(dest)], check=True)
    return str(dest)


def get_current_sha(repo_path: str) -> str:
    result = subprocess.run(
        ["git", "-C", repo_path, "rev-parse", "HEAD"],
        check=True,
        capture_output=True,
        text=True,
    )
    return result.stdout.strip()


def get_changed_files(repo_path: str, since_sha: str) -> list[str]:
    result = subprocess.run(
        ["git", "-C", repo_path, "diff", since_sha, "HEAD", "--name-only"],
        check=True,
        capture_output=True,
        text=True,
    )
    return [f for f in result.stdout.strip().splitlines() if f]
