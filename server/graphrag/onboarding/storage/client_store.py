import json
from dataclasses import asdict, dataclass
from datetime import datetime
from pathlib import Path


@dataclass
class ClientState:
    client_id: str
    graphrag_root: str
    last_indexed_sha: str
    created_at: str = ""

    def __post_init__(self):
        if not self.created_at:
            self.created_at = datetime.utcnow().isoformat()


class ClientStore:
    def __init__(self, path: Path = Path("data/clients.json")):
        self._path = path
        self._path.parent.mkdir(parents=True, exist_ok=True)
        if not self._path.exists():
            self._path.write_text("{}")

    def _load(self) -> dict:
        return json.loads(self._path.read_text())

    def _write(self, data: dict):
        self._path.write_text(json.dumps(data, indent=2))

    def save(self, state: ClientState):
        data = self._load()
        data[state.client_id] = asdict(state)
        self._write(data)

    def get(self, client_id: str) -> ClientState | None:
        data = self._load()
        entry = data.get(client_id)
        if entry is None:
            return None
        return ClientState(**entry)


client_store = ClientStore()
