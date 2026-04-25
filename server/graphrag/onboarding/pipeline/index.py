import subprocess


def run_index(graphrag_root: str):
    subprocess.run(["graphrag", "index", "--root", graphrag_root], check=True)


def run_update(graphrag_root: str):
    subprocess.run(["graphrag", "update", "--root", graphrag_root], check=True)
