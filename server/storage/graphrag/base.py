from abc import ABC, abstractmethod
from pathlib import Path

from storage.graphrag.models import ClientState


class GraphRAGStorage(ABC):
    @abstractmethod
    def save_state(self, state: ClientState) -> None: ...

    @abstractmethod
    def get_state(self, client_id: str) -> ClientState | None: ...

    @abstractmethod
    def get_root(self, client_id: str) -> Path:
        """Returns the local path graphrag should use as its working directory for this client."""
        ...

    @abstractmethod
    def save_artifacts(self, client_id: str, root: Path) -> None:
        """Persist all artifacts (output/, prompts/, settings.yaml, input/, cache/) after indexing completes."""
        ...

    @abstractmethod
    def load_artifacts(self, client_id: str, dest: Path) -> None:
        """Restore artifacts to dest before running update or query. No-op if already local."""
        ...
