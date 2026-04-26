import logging
import subprocess

log = logging.getLogger(__name__)


def _run(cmd: list[str], label: str):
    log.info("Running %s: %s", label, " ".join(cmd))
    result = subprocess.run(cmd, capture_output=True, text=True)
    if result.stdout:
        log.debug("%s stdout:\n%s", label, result.stdout)
    if result.stderr:
        log.debug("%s stderr:\n%s", label, result.stderr)
    if result.returncode != 0:
        log.error("%s failed (exit %d):\n%s", label, result.returncode, result.stderr)
        raise subprocess.CalledProcessError(result.returncode, cmd, result.stdout, result.stderr)
    log.info("%s complete", label)


def run_index(graphrag_root: str):
    _run(["graphrag", "index", "--root", graphrag_root], "graphrag-index")


def run_update(graphrag_root: str):
    _run(["graphrag", "update", "--root", graphrag_root], "graphrag-update")
