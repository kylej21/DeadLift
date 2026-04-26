import logging
import os
import shutil
from pathlib import Path

log = logging.getLogger(__name__)

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
    log.info("Configuring graphrag root: %s", root)

    shutil.copy(_TEMPLATE, root / "settings.yaml")
    log.debug("Copied settings template to %s", root / "settings.yaml")

    env_lines = [f"{k}={os.environ.get(k, '')}" for k in _ENV_KEYS]
    (root / ".env").write_text("\n".join(env_lines) + "\n", encoding="utf-8")
    missing = [k for k in _ENV_KEYS if not os.environ.get(k)]
    if missing:
        log.warning("Missing env vars (will be empty in .env): %s", missing)
    else:
        log.debug("All env vars present: %s", _ENV_KEYS)

    for subdir in ("input", "output"):
        (root / subdir).mkdir(exist_ok=True)
        log.debug("Ensured subdir: %s", root / subdir)

    log.info("Configuration complete for %s", root)
