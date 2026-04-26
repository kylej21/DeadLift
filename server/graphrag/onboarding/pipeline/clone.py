import logging
import subprocess
from pathlib import Path
from urllib.parse import urlparse, urlunparse

log = logging.getLogger(__name__)


def _inject_token(repo_url: str, github_token: str) -> str:
    """Rewrite https://github.com/... to embed OAuth token. Never call this for logging."""
    parsed = urlparse(repo_url)
    if parsed.hostname and "github.com" in parsed.hostname and github_token:
        authed = parsed._replace(netloc=f"x-access-token:{github_token}@{parsed.hostname}")
        return urlunparse(authed)
    return repo_url


def clone_repo(repo_url: str, dest_dir: str, github_token: str = "") -> str:
    dest = Path(dest_dir)
    dest.parent.mkdir(parents=True, exist_ok=True)
    log.info("Cloning %s -> %s", repo_url, dest)  # log original URL, not authed URL
    clone_url = _inject_token(repo_url, github_token)
    subprocess.run(["git", "clone", clone_url, str(dest)], check=True)
    log.info("Clone complete: %s", dest)
    return str(dest)


def pull_repo(repo_path: str, github_token: str = ""):
    """Fetch and pull latest changes. If token provided, temporarily embeds it in remote URL."""
    log.info("Pulling latest changes: %s", repo_path)
    if github_token:
        # Get current remote URL to restore it after pull
        result = subprocess.run(
            ["git", "-C", repo_path, "remote", "get-url", "origin"],
            check=True, capture_output=True, text=True,
        )
        original_url = result.stdout.strip()
        authed_url = _inject_token(original_url, github_token)
        try:
            subprocess.run(["git", "-C", repo_path, "remote", "set-url", "origin", authed_url], check=True)
            subprocess.run(["git", "-C", repo_path, "fetch", "origin"], check=True)
            subprocess.run(["git", "-C", repo_path, "pull"], check=True)
        finally:
            # Always strip token from remote URL, even if pull fails
            subprocess.run(["git", "-C", repo_path, "remote", "set-url", "origin", original_url], check=True)
    else:
        subprocess.run(["git", "-C", repo_path, "fetch", "origin"], check=True)
        subprocess.run(["git", "-C", repo_path, "pull"], check=True)
    log.info("Pull complete: %s", repo_path)


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
