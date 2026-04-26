import json
import logging
import os
import subprocess
from datetime import datetime, timezone
from pathlib import Path

from dotenv import load_dotenv
from fastapi import BackgroundTasks, FastAPI, HTTPException
from pydantic import BaseModel

load_dotenv(Path(__file__).parent / ".env")
_env_mode = os.environ.get("ENV_MODE", "openai")
load_dotenv(Path(__file__).parent / f".env.{_env_mode}")

from .jobs.manager import JobStatus, job_manager
from .pipeline.clone import clone_repo, get_changed_files, get_current_sha
from .pipeline.configure import configure
from .pipeline.index import run_index, run_update
from .pipeline.preprocess import preprocess_all, preprocess_files
from .pipeline.tune import run_prompt_tune
from storage import ClientState, create_storage

logging.basicConfig(
    level=logging.DEBUG,
    format="%(asctime)s %(levelname)-8s %(name)s - %(message)s",
)
log = logging.getLogger(__name__)

app = FastAPI()

storage = create_storage()


class OnboardRequest(BaseModel):
    repo_url: str
    client_id: str


class UpdateRequest(BaseModel):
    client_id: str


def _run_onboard(job_id: str, repo_url: str, client_id: str):
    log.info("[%s] Onboard started: client=%s repo=%s", job_id, client_id, repo_url)
    try:
        job_manager.update_job(job_id, JobStatus.RUNNING, "Cloning repository")
        repo_path = clone_repo(repo_url, str(Path("kb/repos") / client_id))

        job_manager.update_job(job_id, JobStatus.RUNNING, "Preprocessing files")
        graphrag_root = str(storage.get_root(client_id))
        log.info("[%s] graphrag root: %s", job_id, graphrag_root)
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
        log.info("[%s] Onboard complete: client=%s sha=%s", job_id, client_id, sha)
        job_manager.update_job(job_id, JobStatus.COMPLETED, "Knowledge base ready")
    except Exception as exc:
        log.exception("[%s] Onboard failed: %s", job_id, exc)
        job_manager.update_job(job_id, JobStatus.FAILED, str(exc))


def _run_update(job_id: str, client_id: str):
    log.info("[%s] Update started: client=%s", job_id, client_id)
    try:
        state = storage.get_state(client_id)
        if state is None:
            log.error("[%s] Client not found: %s", job_id, client_id)
            job_manager.update_job(job_id, JobStatus.FAILED, f"Client {client_id} not found")
            return

        repo_path = str(Path("kb/repos") / client_id)
        graphrag_root = str(storage.get_root(client_id))
        log.info("[%s] Fetching repo: %s", job_id, repo_path)

        job_manager.update_job(job_id, JobStatus.RUNNING, "Fetching latest changes")
        subprocess.run(["git", "-C", repo_path, "fetch", "origin"], check=True)
        subprocess.run(["git", "-C", repo_path, "pull"], check=True)

        changed_files = get_changed_files(repo_path, state.last_indexed_sha)
        if not changed_files:
            log.info("[%s] No changes detected, skipping update", job_id)
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
        log.info("[%s] Update complete: client=%s sha=%s", job_id, client_id, sha)
        job_manager.update_job(job_id, JobStatus.COMPLETED, "Knowledge base updated")
    except Exception as exc:
        log.exception("[%s] Update failed: %s", job_id, exc)
        job_manager.update_job(job_id, JobStatus.FAILED, str(exc))


@app.post("/onboard")
def onboard(req: OnboardRequest, background_tasks: BackgroundTasks):
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


RCA_DIR = Path(__file__).parent / "rca_files"
RCA_DIR.mkdir(exist_ok=True)


class RCACreateRequest(BaseModel):
    org_id: str
    message_id: str
    error_class: str
    raw_payload: str
    fixed_payload: str
    analysis: str


@app.post("/rca/create")
def rca_create(req: RCACreateRequest):
    filename = f"{req.org_id}_{req.message_id}.json"
    path = RCA_DIR / filename
    data = {
        "org_id": req.org_id,
        "message_id": req.message_id,
        "error_class": req.error_class,
        "raw_payload": req.raw_payload,
        "fixed_payload": req.fixed_payload,
        "analysis": req.analysis,
        "created_at": datetime.now(timezone.utc).isoformat(),
    }
    path.write_text(json.dumps(data, indent=2))
    log.info("rca: wrote %s", filename)
    return {"ok": True, "file": filename}


@app.get("/rca/get/{org_id}")
def rca_get(org_id: str):
    reports = []
    for path in sorted(RCA_DIR.glob(f"{org_id}_*.json"), reverse=True):
        try:
            reports.append(json.loads(path.read_text()))
        except Exception as exc:
            log.warning("rca: failed to read %s: %s", path.name, exc)
    return reports
