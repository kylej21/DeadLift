""" 
Define all MCP tools for LQL, Graphrag and Bigquery 
"""

import os 
from datetime import datetime, timedelta, timezone 
from dotenv import load_dotenv 
from fastmcp import FastMCP 
from google.cloud import logging as gcp_logging 

load_dotenv() 

mcp = FastMCP() 

log_client = gcp_logging.Client()

@mcp.tool() 
def fetch_gcp_logs(resource_type: str = "cloud_run_revision",
    severity: str = "ERROR",
    hours_back: int = 1, 
    max_entries: int = 50, 
) -> list[dict]:  
    """ 
    Fetch Recent GCP logs filtered by resource type, severity, and time range. 
    """
    cutoff = datetime.now(timezone.utc) - timedelta(hours=hours_back)
    filter_str = (
        f'resource.type="{resource_type}" '
        f'severity>="{severity}" '
        f'timestamp>="{cutoff.isoformat()}"'
    )

    entries = [] 
    for entry in log_client.list_entries(
        filter_=filter_str, 
        max_results=max_entries, 
        order_by=gcp_logging.DESCENDING,
    ): 
        entries.append({
            "timestamp": entry.timestamp.isoformat() if entry.timestamp else None,
            "severity": entry.severity,
            "resource_type": entry.resource.type if entry.resource else None,
            "resource_labels": dict(entry.resource.labels) if entry.resource else {},
            "message": str(entry.payload) if entry.payload else "",
            "log_name": entry.log_name,
        })
    return entries

@mcp.tool() 
def list_log_resource_types() -> list[str]:
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