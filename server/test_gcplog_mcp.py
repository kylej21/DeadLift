import asyncio
import json
import os
from dotenv import load_dotenv
import google.generativeai as genai
from fastmcp import Client

load_dotenv()
genai.configure(api_key=os.environ["GEMINI_API_KEY"])


async def diagnose_logs(user_query: str):
    async with Client("mcp_server.py") as mcp_client:
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
                "You are a cloud infrastructure diagnostics expert. "
                "Use the available tools to fetch GCP logs, then analyze them "
                "to identify root causes, error patterns, and actionable fixes. "
                "Always state which resource type and time window you fetched, "
                "then summarize findings clearly."
            ),
        )

        chat = model.start_chat()
        response = chat.send_message(user_query)

        # Agentic tool-call loop
        while True:
            parts = response.candidates[0].content.parts
            fn_parts = [p for p in parts if hasattr(p, "function_call") and p.function_call.name]

            if not fn_parts:
                print("\n=== Diagnosis ===\n")
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
    asyncio.run(diagnose_logs(
        "Fetch the last 2 ERROR and CRITICAL logs from Cloud Run "
        "and give me a diagnosis with suggested fixes."
    ))