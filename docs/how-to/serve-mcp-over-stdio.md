# How to serve MCP over Standard IO (Stdio)

`idx` natively offers a strict local Standard Input/Output server layer for standard MCP protocols. Stdio is specifically tailored for securely provisioning knowledge base contexts cleanly directly into process-managed tools (such as Desktop LLM Agents like Cursor and Claude) without managing complex network surfaces.

## Prerequisites

1. Ensure the CLI is built locally into a permanent persistent path.
2. Ensure you have an active locally running Ollama environment.

## Configuring your Agent

You do not run `idx serve stdio` manually in your day-to-day workflow. Instead, you integrate the binary directly into your AI Assistant's managed tool context. For example, within the local Claude Desktop configuration (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "idx-notes": {
      "command": "/absolute/path/to/bin/idx",
      "args": [
        "serve",
        "stdio",
        "--dir",
        "/absolute/path/to/my/markdown/notes"
      ]
    }
  }
}
```

The desktop program handles executing the binary passing the STDIN inputs while decoding the STDOUT JSON payloads returned from your embedded RAG structure, establishing a permanent integration of your Markdown data directly with your prompt iterations.
