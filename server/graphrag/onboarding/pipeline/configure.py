import os
import shutil
from pathlib import Path

_ONBOARDING_ROOT = Path(__file__).parent.parent
_TEMPLATE = _ONBOARDING_ROOT / "settings_template.yaml"

_ENV_KEYS = [
    "GRAPHRAG_API_KEY",
    "GRAPHRAG_LLM_BASE_URL",
    "GRAPHRAG_MODEL",
    "GRAPHRAG_EMBEDDING_BASE_URL",
    "GRAPHRAG_EMBEDDING_MODEL",
]


def configure(graphrag_root: str):
    root = Path(graphrag_root)
    root.mkdir(parents=True, exist_ok=True)
    shutil.copy(_TEMPLATE, root / "settings.yaml")
    env_lines = [f"{k}={os.environ.get(k, '')}" for k in _ENV_KEYS]
    (root / ".env").write_text("\n".join(env_lines) + "\n", encoding="utf-8")
    for subdir in ("input", "output"):
        (root / subdir).mkdir(exist_ok=True)
