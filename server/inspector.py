"""
Unified MCP inspector. Persistent REPL across BigQuery, GCP logs, and GraphRAG.
Run: python inspect.py
"""

import asyncio
import os
import sys
from pathlib import Path
from dotenv import load_dotenv
from fastmcp import Client
from fastmcp.client.transports import PythonStdioTransport

from llm_tool_agent import run_tool_agent, mcp_tools_to_openai_tools

load_dotenv()

SERVER = str(Path(__file__).parent / "mcp_tools.py")
E2E_WORKSPACE = str(Path(__file__).parent / "e2e-test")

SYSTEM_PROMPT = """You are an infrastructure and data analyst with access to three capabilities via tools:

1. BigQuery — fetch and analyze rows from BigQuery tables (use bigquery_last_n_query).
2. GCP logs — fetch Cloud Run / GCE logs to diagnose errors and patterns.
3. GraphRAG — query a knowledge graph built from codebases and incident data.
   Always start with method='local' for entity-level questions (errors, handlers,
   configs, services). Only use method='global' for system-wide architecture or themes.

Pick the right tool based on the user's question. You may chain tools across turns —
e.g., spot an error in GCP logs, then look it up in GraphRAG, then check BigQuery for
related records. Always state which tool, dataset/resource, and parameters you used,
then summarize findings clearly.
"""


async def repl():
    transport = PythonStdioTransport(
        script_path=SERVER,
        python_cmd=sys.executable,
        env={**os.environ, "GRAPH_RAG_ROOT": E2E_WORKSPACE},
    )

    async with Client(transport) as mcp_client:
        mcp_tools = await mcp_client.list_tools()
        tools = mcp_tools_to_openai_tools(mcp_tools)

        print(f"Loaded {len(tools)} tools: {', '.join(t['function']['name'] for t in tools)}")
        print("Type a question, or '/reset' to clear history, '/exit' to quit.\n")

        messages = [{"role": "system", "content": SYSTEM_PROMPT}]

        while True:
            try:
                query = input("» ").strip()
            except (EOFError, KeyboardInterrupt):
                print()
                break

            if not query:
                continue
            if query == "/exit":
                break
            if query == "/reset":
                messages = [{"role": "system", "content": SYSTEM_PROMPT}]
                print("[history cleared]\n")
                continue

            try:
                await run_tool_agent(
                    mcp_client=mcp_client,
                    user_query=query,
                    messages=messages,
                    tools=tools,
                )
            except Exception as e:
                print(f"[error: {e}]\n")


if __name__ == "__main__":
    asyncio.run(repl())
