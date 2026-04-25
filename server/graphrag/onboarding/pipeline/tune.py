import subprocess
from pathlib import Path


def run_prompt_tune(graphrag_root: str):
    prompts_dir = str(Path(graphrag_root).resolve() / "prompts")
    subprocess.run(
        [
            "graphrag", "prompt-tune",
            "--root", graphrag_root,
            "--output", prompts_dir,
            "--selection-method", "top",
        ],
        check=True,
    )
