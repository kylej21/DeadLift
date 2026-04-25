import os

from storage.graphrag.base import GraphRAGStorage
from storage.graphrag.local import LocalGraphRAGStorage
from storage.graphrag.models import ClientState


def create_storage() -> GraphRAGStorage:
    backend = os.environ.get("STORAGE_BACKEND", "local")
    if backend == "local":
        return LocalGraphRAGStorage()
    # Future backends (e.g. "gcs") will be added here.
    raise ValueError(f"Unknown storage backend: {backend!r}")


__all__ = ["GraphRAGStorage", "LocalGraphRAGStorage", "ClientState", "create_storage"]
