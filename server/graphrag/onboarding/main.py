import subprocess
from pathlib import Path

from dotenv import load_dotenv
from fastapi import BackgroundTasks, FastAPI, HTTPException

import os
load_dotenv(Path(__file__).parent / ".env")
_env_mode = os.environ.get("ENV_MODE", "openai")
load_dotenv(Path(__file__).parent / f".env.{_env_mode}")
from pydantic import BaseModel

from .jobs.manager import JobStatus, job_manager
from .pipeline.clone import clone_repo, get_changed_files, get_current_sha
from .pipeline.configure import configure
from .pipeline.index import run_index, run_update
from .pipeline.preprocess import preprocess_all, preprocess_files
from .pipeline.tune import run_prompt_tune
from storage import ClientState, create_storage

app = FastAPI()

storage = create_storage()


class OnboardRequest(BaseModel):
    repo_url: str
    client_id: str


class UpdateRequest(BaseModel):
    client_id: str


def _run_onboard(job_id: str, repo_url: str, client_id: str):
    try:
        job_manager.update_job(job_id, JobStatus.RUNNING, "Cloning repository")
        repo_path = clone_repo(repo_url, str(Path("kb/repos") / client_id))

        job_manager.update_job(job_id, JobStatus.RUNNING, "Preprocessing files")
        graphrag_root = str(storage.get_root(client_id))
        configure(graphrag_root)
        preprocess_all(repo_path, str(Path(graphrag_root) / "input"))

        job_manager.update_job(job_id, JobStatus.RUNNING, "Tuning prompts")
        run_prompt_tune(graphrag_root)

        job_manager.update_job(job_id, JobStatus.RUNNING, "Indexing knowledge base")
        run_index(graphrag_root)
        storage.save_artifacts(client_id, Path(graphrag_root))

        sha = get_current_sha(repo_path)
        storage.save_state(ClientState(
            client_id=client_id,
            last_indexed_sha=sha,
        ))

        job_manager.update_job(job_id, JobStatus.COMPLETED, "Knowledge base ready")
    except Exception as exc:
        job_manager.update_job(job_id, JobStatus.FAILED, str(exc))


def _run_update(job_id: str, client_id: str):
    try:
        state = storage.get_state(client_id)
        if state is None:
            job_manager.update_job(job_id, JobStatus.FAILED, f"Client {client_id} not found")
            return

        repo_path = str(Path("kb/repos") / client_id)
        graphrag_root = str(storage.get_root(client_id))

        job_manager.update_job(job_id, JobStatus.RUNNING, "Fetching latest changes")
        subprocess.run(["git", "-C", repo_path, "fetch", "origin"], check=True)
        subprocess.run(["git", "-C", repo_path, "pull"], check=True)

        changed_files = get_changed_files(repo_path, state.last_indexed_sha)
        if not changed_files:
            job_manager.update_job(job_id, JobStatus.COMPLETED, "No changes detected")
            return

        job_manager.update_job(job_id, JobStatus.RUNNING, "Preprocessing changed files")
        preprocess_files(repo_path, str(Path(graphrag_root) / "input"), changed_files)

        job_manager.update_job(job_id, JobStatus.RUNNING, "Updating knowledge base")
        run_update(graphrag_root)
        storage.save_artifacts(client_id, Path(graphrag_root))

        sha = get_current_sha(repo_path)
        state.last_indexed_sha = sha
        storage.save_state(state)

        job_manager.update_job(job_id, JobStatus.COMPLETED, "Knowledge base updated")
    except Exception as exc:
        job_manager.update_job(job_id, JobStatus.FAILED, str(exc))


@app.post("/onboard")
def onboard(req: OnboardRequest, background_tasks: BackgroundTasks):
    if storage.client_exists(req.client_id):
        raise HTTPException(status_code=409, detail=f"Client {req.client_id} already exists. Use /update to add new changes.")
    job = job_manager.create_job(req.client_id)
    background_tasks.add_task(_run_onboard, job.job_id, req.repo_url, req.client_id)
    return {"job_id": job.job_id}


@app.post("/update")
def update(req: UpdateRequest, background_tasks: BackgroundTasks):
    state = storage.get_state(req.client_id)
    if state is None:
        raise HTTPException(status_code=404, detail=f"Client {req.client_id} not found")
    job = job_manager.create_job(req.client_id)
    background_tasks.add_task(_run_update, job.job_id, req.client_id)
    return {"job_id": job.job_id}


@app.get("/status/{job_id}")
def status(job_id: str):
    job = job_manager.get_job(job_id)
    if job is None:
        raise HTTPException(status_code=404, detail="Job not found")
    return {
        "job_id": job.job_id,
        "status": job.status,
        "message": job.message,
        "client_id": job.client_id,
    }
