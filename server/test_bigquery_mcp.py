"""
Gemini-driven BigQuery inspector via MCP.
Run: python test_bigquery_mcp.py
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


async def inspect_bigquery(user_query: str):
    transport = PythonStdioTransport(script_path=SERVER, python_cmd=sys.executable)
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
                "You are a data analyst with access to BigQuery. "
                "Use the bigquery_last_n_query tool to fetch rows from the requested table, "
                "then summarize what you observe: the schema, notable values, patterns, or anomalies. "
                "Always state the dataset and table you queried."
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
                result_data = tool_result.content[0].text if tool_result.content else "[]"

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
    asyncio.run(inspect_bigquery(
        "Fetch 5 rows from the deadlift dataset, success_logs table, "
        "and tell me what the data looks like."
    ))
