import json
from dataclasses import asdict
from pathlib import Path

from storage.graphrag.base import GraphRAGStorage
from storage.graphrag.models import ClientState


class LocalGraphRAGStorage(GraphRAGStorage):
    def __init__(
        self,
        base_dir: Path = Path("data/graphrag"),
        state_file: Path = Path("data/clients.json"),
    ):
        self._base_dir = base_dir
        self._path = state_file
        self._path.parent.mkdir(parents=True, exist_ok=True)
        if not self._path.exists():
            self._path.write_text("{}")

    def _load(self) -> dict:
        return json.loads(self._path.read_text())

    def _write(self, data: dict):
        self._path.write_text(json.dumps(data, indent=2))

    def get_root(self, client_id: str) -> Path:
        """Returns the local path graphrag should use as its working directory for this client."""
        root = self._base_dir / client_id
        root.mkdir(parents=True, exist_ok=True)
        return root

    def save_artifacts(self, client_id: str, root: Path) -> None:
        """Persist all artifacts (output/, prompts/, settings.yaml, input/, cache/) after indexing completes."""

    def load_artifacts(self, client_id: str, dest: Path) -> None:
        """Restore artifacts to dest before running update or query. No-op if already local."""

    def save_state(self, state: ClientState) -> None:
        data = self._load()
        data[state.client_id] = asdict(state)
        self._write(data)

    def get_state(self, client_id: str) -> ClientState | None:
        data = self._load()
        entry = data.get(client_id)
        if entry is None:
            return None
        return ClientState(**entry)
