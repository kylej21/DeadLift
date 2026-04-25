from dataclasses import dataclass
from datetime import datetime


@dataclass
class ClientState:
    client_id: str
    last_indexed_sha: str
    created_at: str = ""

    def __post_init__(self):
        if not self.created_at:
            self.created_at = datetime.utcnow().isoformat()
