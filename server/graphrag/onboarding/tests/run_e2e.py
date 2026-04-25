#!/usr/bin/env python3
"""
End-to-end test for the GraphRAG onboarding pipeline.
Run from anywhere:
    python server/graphrag/onboarding/tests/run_e2e.py --provider openai
    python server/graphrag/onboarding/tests/run_e2e.py --provider local

Or with a real repo:
    python server/graphrag/onboarding/tests/run_e2e.py --provider openai --repo https://github.com/... --client-id my-client

Provider settings are loaded from .env.<provider> in the onboarding root.
GRAPHRAG_API_KEY must be set in .env (for openai) or is set to "ollama" in .env.local.
"""

import argparse, os, shutil, subprocess, sys, time
from datetime import datetime
from pathlib import Path
from dotenv import load_dotenv

# --- path setup BEFORE any local imports ---
ONBOARDING_ROOT = Path(__file__).parent.parent.resolve()
SERVER_ROOT = ONBOARDING_ROOT.parent.parent
FIXTURES_DIR = Path(__file__).parent / "fixtures"
sys.path.insert(0, str(SERVER_ROOT))

_REQUIRED_ENV = [
    "GRAPHRAG_API_KEY",
    "GRAPHRAG_LLM_BASE_URL",
    "GRAPHRAG_MODEL",
    "GRAPHRAG_EMBEDDING_BASE_URL",
    "GRAPHRAG_EMBEDDING_MODEL",
]


def stage(name, fn):
    print(f"  [ RUN ] {name}", end="", flush=True)
    t = time.time()
    try:
        result = fn()
        elapsed = time.time() - t
        print(f"\r  [PASS] {name:<25} {elapsed:.1f}s" + (f"  {result}" if result else ""))
        return result
    except Exception as e:
        elapsed = time.time() - t
        print(f"\r  [FAIL] {name:<25} {elapsed:.1f}s")
        print(f"         {type(e).__name__}: {e}")
        sys.exit(1)


def check_outputs(graphrag_root):
    root = Path(graphrag_root)
    checks = [
        root / "settings.yaml",
        root / "prompts" / "extract_graph.txt",
        root / "input",
    ]
    output_parquets = [
        "entities.parquet",
        "relationships.parquet",
        "communities.parquet",
    ]
    for f in checks:
        assert f.exists(), f"Missing: {f}"

    input_files = list((root / "input").glob("*.txt"))
    assert len(input_files) > 0, "No input files found"

    output_dir = root / "output"
    found_parquets = []
    for parquet_name in output_parquets:
        matches = list(output_dir.rglob(parquet_name)) if output_dir.exists() else []
        if matches:
            size_kb = matches[0].stat().st_size // 1024
            found_parquets.append(f"{parquet_name} {size_kb}KB")
            assert matches[0].stat().st_size > 0, f"{parquet_name} is empty"

    return f"{len(input_files)} input files, " + ", ".join(found_parquets)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--provider", choices=["openai", "local"], required=True,
                        help="LLM provider: 'openai' or 'local' (Ollama)")
    parser.add_argument("--repo", default=None)
    parser.add_argument("--client-id", default="e2e-test")
    args = parser.parse_args()

    # Load provider file first, then .env for secrets (load_dotenv won't override already-set vars)
    load_dotenv(ONBOARDING_ROOT / f".env.{args.provider}")
    load_dotenv(ONBOARDING_ROOT / ".env")

    missing = [k for k in _REQUIRED_ENV if not os.environ.get(k)]
    if missing:
        print(f"Missing env vars: {', '.join(missing)}")
        print(f"Check .env.{args.provider} and .env in {ONBOARDING_ROOT}")
        sys.exit(1)

    print(f"Provider: {args.provider}")
    for k in _REQUIRED_ENV:
        print(f"  {k} = {os.environ[k]}")
    print()

    # create output dir and chdir BEFORE importing modules with singletons
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    output_dir = ONBOARDING_ROOT / "tests" / "e2e_output" / timestamp
    output_dir.mkdir(parents=True, exist_ok=True)
    os.chdir(output_dir)
    print(f"Output directory: {output_dir}\n")

    # now safe to import (singletons will use paths relative to output_dir)
    from graphrag.onboarding.jobs.manager import job_manager, JobStatus
    from graphrag.onboarding.main import _run_onboard, _run_update

    repo_url = args.repo
    remote_dir = None
    if repo_url is None:
        remote_dir = output_dir / "sample_remote"
        shutil.copytree(FIXTURES_DIR / "sample_repo", remote_dir)
        subprocess.run(["git", "init"], cwd=remote_dir, check=True, capture_output=True)
        subprocess.run(["git", "add", "."], cwd=remote_dir, check=True, capture_output=True)
        subprocess.run(["git", "-c", "user.email=test@test.com", "-c", "user.name=Test",
                        "commit", "-m", "initial"], cwd=remote_dir, check=True, capture_output=True)
        repo_url = remote_dir.as_uri()
        print(f"Using fixture repo: {repo_url}\n")

    client_id = args.client_id
    total_start = time.time()

    print("=== Onboard ===")
    job = job_manager.create_job(client_id)
    stage("Full onboard pipeline", lambda: _run_onboard(job.job_id, repo_url, client_id))

    final_job = job_manager.get_job(job.job_id)
    if final_job.status != JobStatus.COMPLETED:
        print(f"\n[FAIL] Job ended with status {final_job.status}: {final_job.message}")
        sys.exit(1)

    from storage.graphrag import LocalGraphRAGStorage
    _store = LocalGraphRAGStorage()
    state = _store.get_state(client_id)
    stage("Output structure check", lambda: check_outputs(_store.get_root(client_id)))
    print(f"       SHA: {state.last_indexed_sha}")

    if remote_dir:
        print("\n=== Update (simulated change) ===")
        handler = remote_dir / "src" / "handler.py"
        with open(handler, "a") as f:
            f.write("\n# updated by e2e test\n")
        subprocess.run(["git", "add", "."], cwd=remote_dir, check=True, capture_output=True)
        subprocess.run(["git", "-c", "user.email=test@test.com", "-c", "user.name=Test",
                        "commit", "-m", "update handler"], cwd=remote_dir, check=True, capture_output=True)

        update_job = job_manager.create_job(client_id)
        stage("Update pipeline", lambda: _run_update(update_job.job_id, client_id))
        final_update = job_manager.get_job(update_job.job_id)
        if final_update.status != JobStatus.COMPLETED:
            print(f"\n[FAIL] Update job ended with status {final_update.status}: {final_update.message}")
            sys.exit(1)
        new_state = _store.get_state(client_id)
        print(f"       SHA updated: {state.last_indexed_sha} → {new_state.last_indexed_sha}")

    total = time.time() - total_start
    print(f"\nAll checks passed. Total: {total:.1f}s")
    print(f"Output at: {output_dir}")


if __name__ == "__main__":
    main()
