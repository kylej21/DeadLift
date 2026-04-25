""" 
Define all MCP tools for LQL, Graphrag and Bigquery 
"""

import os
import subprocess
from dotenv import load_dotenv
from fastmcp import FastMCP
from google.cloud import logging as gcp_logging
from google.cloud import bigquery

load_dotenv() 

mcp = FastMCP() 

log_client = gcp_logging.Client()
bigquery_client = bigquery.Client(project="your-project-id")

# graph rag root directory, subject to change based on where the graph rag lives 
GRAPH_RAG_ROOT = os.path.join(os.path.dirname(__file__), "graph_rag_workspace")

@mcp.tool()
def fetch_gcp_logs(
    resource_type: str = "cloud_run_revision",
    severity: str = "ERROR",
    page_size: int = 25,
    page_token: str | None = None,
) -> dict:
    """
    Fetch GCP logs filtered by resource type and severity, with pagination.
    Returns the most recent `page_size` logs and a next_page_token.
    If the relevant log isn't in the returned batch, call again with next_page_token
    to page further back in time.
    """
    filter_str = (
        f'resource.type="{resource_type}" '
        f'severity>="{severity}"'
    )

    iterator = log_client.list_entries(
        filter_=filter_str,
        max_results=page_size,
        order_by=gcp_logging.DESCENDING,
        page_token=page_token,
    )

    entries = []
    for entry in iterator:
        payload = entry.payload
        if isinstance(payload, dict):
            parsed_payload = payload
        else:
            parsed_payload = str(payload) if payload else ""

        entries.append({
            "insert_id": entry.insert_id,
            "timestamp": entry.timestamp.isoformat() if entry.timestamp else None,
            "severity": entry.severity,
            "log_name": entry.log_name,
            "resource_type": entry.resource.type if entry.resource else None,
            "resource_labels": dict(entry.resource.labels) if entry.resource else {},
            "payload": parsed_payload,
        })
        if len(entries) >= page_size:
            break

    return {
        "entries": entries,
        "count": len(entries),
        "next_page_token": getattr(iterator, "next_page_token", None),
    }

@mcp.tool() 
def gcp_list_log_resource_types() -> list[str]:
    """List common GCP monitored resource types for log filtering."""
    return [
        "gce_instance",
        "cloud_run_revision",
        "k8s_container",
        "cloud_function",
        "appengine_app",
        "cloudsql_database",
        "gcs_bucket",
    ]

@mcp.tool()
def graph_rag_query(
    query: str,
    method: str = "local",
    response_type: str = "Multiple Paragraphs",
    community_level: int = 2,
) -> dict:
    """
    Query the GraphRAG knowledge graph built from the codebase/incident data.
    Use method='local' or 'drift' for specific entity/error questions (best for dead letter diagnosis).
    Use method='global' for broad questions about the whole dataset.
    """
    cmd = [
        "graphrag", "query",
        "--root", GRAPH_RAG_ROOT,
        "--method", method,
        "--response-type", response_type,
        "--community-level", str(community_level),
        query,
    ]

    result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)

    if result.returncode != 0:
        return {
            "success": False,
            "error": result.stderr.strip() or "graphrag query failed with no stderr output",
        }

    return {
        "success": True,
        "method": method,
        "response": result.stdout.strip(),
    }

@mcp.tool()
def bigquery_last_n_query(
    dataset: str,
    table: str,
    n: int = 5,
    order_by: str | None = None,
) -> list[dict]:
    """
    Fetch N rows from a BigQuery table (equivalent to:
    bq query --nouse_legacy_sql 'SELECT * FROM `project.dataset.table` LIMIT n').

    Args:
        dataset: BigQuery dataset name (e.g. "deadlift")
        table: BigQuery table name (e.g. "success_logs")
        n: Number of rows to return (default 5, max 100)
        order_by: Optional column to sort by descending (e.g. a timestamp column)
    """
    n = min(n, 100) # safety cap of 100 prev entries 
    order_clause = f"ORDER BY {order_by} DESC" if order_by else ""
    query = f"SELECT * FROM `{bigquery_client.project}.{dataset}.{table}` {order_clause} LIMIT {n}"
    results = bigquery_client.query(query)
    return [dict(row) for row in results]

if __name__ == "__main__":
    mcp.run() 