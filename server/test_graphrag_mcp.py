"""
Gemini-driven GraphRAG inspector via MCP.
Run: python test_graphrag_mcp.py
"""

import asyncio
import os
import sys
from pathlib import Path
from dotenv import load_dotenv
import google.generativeai as genai
from fastmcp import Client
from fastmcp.client.transports import PythonStdioTransport

load_dotenv()
genai.configure(api_key=os.environ["GEMINI_API_KEY"])

SERVER = str(Path(__file__).parent / "mcp_server.py")
E2E_WORKSPACE = str(Path(__file__).parent / "e2e-test")


async def inspect_graphrag(user_query: str):
    transport = PythonStdioTransport(
        script_path=SERVER,
        python_cmd=sys.executable,
        env={**os.environ, "GRAPH_RAG_ROOT": E2E_WORKSPACE},
    )
    async with Client(transport) as mcp_client:
        mcp_tools = await mcp_client.list_tools()

        gemini_tools = []
        for t in mcp_tools:
            properties = {}
            for k, v in (t.inputSchema.get("properties") or {}).items():
                prop_type = v.get("type", "string")
                type_map = {
                    "string": genai.protos.Type.STRING,
                    "integer": genai.protos.Type.INTEGER,
                    "boolean": genai.protos.Type.BOOLEAN,
                    "number": genai.protos.Type.NUMBER,
                }
                properties[k] = genai.protos.Schema(
                    type=type_map.get(prop_type, genai.protos.Type.STRING),
                    description=v.get("description", ""),
                )

            gemini_tools.append(
                genai.protos.Tool(
                    function_declarations=[
                        genai.protos.FunctionDeclaration(
                            name=t.name,
                            description=t.description,
                            parameters=genai.protos.Schema(
                                type=genai.protos.Type.OBJECT,
                                properties=properties,
                            ),
                        )
                    ]
                )
            )

        model = genai.GenerativeModel(
            model_name="gemini-2.5-flash",
            tools=gemini_tools,
            system_instruction=(
                "You are an expert at querying knowledge graphs built from codebases and incident data. "
                "Use the graph_rag_query tool to answer questions. "
                "IMPORTANT: always start with method='local' — it searches individual entities like errors, "
                "handlers, configs, and services directly. Only switch to method='global' if the question "
                "is explicitly about the overall system architecture or high-level themes. "
                "The knowledge graph contains entities such as error types (ValidationError, ProcessingError), "
                "dead letter topics, retry policies, message handlers, and Pub/Sub topics. "
                "Always state which method you used and summarize findings clearly."
            ),
        )

        chat = model.start_chat()
        response = chat.send_message(user_query)

        while True:
            parts = response.candidates[0].content.parts
            fn_parts = [p for p in parts if hasattr(p, "function_call") and p.function_call.name]

            if not fn_parts:
                print("\n=== Analysis ===\n")
                print(response.text)
                break

            for part in fn_parts:
                fn = part.function_call
                print(f"[→ calling {fn.name}({dict(fn.args)})]")

                tool_result = await mcp_client.call_tool(fn.name, dict(fn.args))
                result_data = tool_result.content[0].text if tool_result.content else "{}"

                response = chat.send_message(
                    genai.protos.Content(
                        parts=[
                            genai.protos.Part(
                                function_response=genai.protos.FunctionResponse(
                                    name=fn.name,
                                    response={"result": result_data},
                                )
                            )
                        ]
                    )
                )


if __name__ == "__main__":
    asyncio.run(inspect_graphrag(
        "What error types exist in this codebase and how does the system handle dead letter messages?"
    ))
