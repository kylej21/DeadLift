import logging
import subprocess
from pathlib import Path

log = logging.getLogger(__name__)


def run_prompt_tune(graphrag_root: str):
    prompts_dir = str(Path(graphrag_root).resolve() / "prompts")
    log.info("Starting prompt tuning: root=%s output=%s", graphrag_root, prompts_dir)
    result = subprocess.run(
        [
            "graphrag", "prompt-tune",
            "--root", graphrag_root,
            "--output", prompts_dir,
            "--selection-method", "top",
        ],
        capture_output=True,
        text=True,
    )
    if result.stdout:
        log.debug("prompt-tune stdout:\n%s", result.stdout)
    if result.stderr:
        log.debug("prompt-tune stderr:\n%s", result.stderr)
    if result.returncode != 0:
        log.error("prompt-tune failed (exit %d):\n%s", result.returncode, result.stderr)
        raise subprocess.CalledProcessError(result.returncode, result.args, result.stdout, result.stderr)
    log.info("Prompt tuning complete")
