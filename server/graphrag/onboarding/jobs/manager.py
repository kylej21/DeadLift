from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from uuid import uuid4


class JobStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"


@dataclass
class Job:
    job_id: str
    client_id: str
    status: JobStatus
    message: str
    created_at: datetime = field(default_factory=datetime.utcnow)


class JobManager:
    def __init__(self):
        self._jobs: dict[str, Job] = {}

    def create_job(self, client_id: str) -> Job:
        job = Job(
            job_id=str(uuid4()),
            client_id=client_id,
            status=JobStatus.PENDING,
            message="Job queued",
        )
        self._jobs[job.job_id] = job
        return job

    def update_job(self, job_id: str, status: JobStatus, message: str):
        job = self._jobs[job_id]
        job.status = status
        job.message = message

    def get_job(self, job_id: str) -> Job | None:
        return self._jobs.get(job_id)


job_manager = JobManager()
