""" 
Define all MCP tools for LQL, Graphrag and Bigquery 
"""

import os 
from dotenv import load_dotenv 
from fastmcp import FastMCP 
from google.cloud import logging as gcp_logging 

load_dotenv() 

mcp = FastMCP() 

log_client = gcp_logging.Client()

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
def tool3(): pass 

@mcp.tool()
def tool4(): pass

if __name__ == "__main__":
    mcp.run() 