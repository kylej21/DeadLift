"""
Shared tool-call loop for MCP + vLLM (OpenAI-compatible) clients.
"""

import json
import os
from openai import OpenAI

llm = OpenAI(
    base_url=os.getenv("VLLM_BASE_URL", "http://127.0.0.1:8000/v1"),
    api_key=os.getenv("VLLM_API_KEY", "dummy"),
)
MODEL = os.getenv("VLLM_MODEL", "incident-gemma")


def mcp_tools_to_openai_tools(mcp_tools):
    return [
        {
            "type": "function",
            "function": {
                "name": t.name,
                "description": t.description or "",
                "parameters": t.inputSchema or {"type": "object", "properties": {}},
            },
        }
        for t in mcp_tools
    ]


async def run_tool_agent(
    mcp_client,
    user_query: str,
    messages: list,           # <-- caller owns history now
    tools: list,              # <-- pre-converted, passed in
    max_tool_rounds: int = 8,
    verbose: bool = True,
):
    """
    Append user_query to messages, run the tool loop, return the final
    assistant text. messages is mutated in place so the caller keeps history.
    """
    messages.append({"role": "user", "content": user_query})

    for _ in range(max_tool_rounds):
        response = llm.chat.completions.create(
            model=MODEL,
            messages=messages,
            tools=tools,
            tool_choice="auto",
            parallel_tool_calls=False,
        )
        message = response.choices[0].message

        if not message.tool_calls:
            messages.append(message)
            if verbose:
                print(f"\n=== Response ===\n{message.content or ''}\n")
            return message.content

        messages.append(message)

        for call in message.tool_calls:
            try:
                args = json.loads(call.function.arguments or "{}")
            except json.JSONDecodeError:
                args = {}

            if verbose:
                print(f"[→ {call.function.name}({args})]")

            tool_result = await mcp_client.call_tool(call.function.name, args)
            result_text = tool_result.content[0].text if tool_result.content else ""

            messages.append({
                "role": "tool",
                "tool_call_id": call.id,
                "name": call.function.name,
                "content": result_text,
            })

    raise RuntimeError(f"Exceeded {max_tool_rounds} tool rounds.")