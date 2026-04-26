import logging
import subprocess
from pathlib import Path

log = logging.getLogger(__name__)


def clone_repo(repo_url: str, dest_dir: str) -> str:
    dest = Path(dest_dir)
    dest.parent.mkdir(parents=True, exist_ok=True)
    log.info("Cloning %s -> %s", repo_url, dest)
    subprocess.run(["git", "clone", repo_url, str(dest)], check=True)
    log.info("Clone complete: %s", dest)
    return str(dest)


def get_current_sha(repo_path: str) -> str:
    result = subprocess.run(
        ["git", "-C", repo_path, "rev-parse", "HEAD"],
        check=True,
        capture_output=True,
        text=True,
    )
    sha = result.stdout.strip()
    log.debug("HEAD sha for %s: %s", repo_path, sha)
    return sha


def get_changed_files(repo_path: str, since_sha: str) -> list[str]:
    log.info("Getting changed files since %s in %s", since_sha, repo_path)
    result = subprocess.run(
        ["git", "-C", repo_path, "diff", since_sha, "HEAD", "--name-only"],
        check=True,
        capture_output=True,
        text=True,
    )
    files = [f for f in result.stdout.strip().splitlines() if f]
    log.info("Found %d changed file(s): %s", len(files), files)
    return files
