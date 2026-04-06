# How to serve MCP over HTTP (SSE)

`idx` provides a Server-Sent Events (SSE) server for remote or networked clients compatible with the Model Context Protocol (MCP) HTTP standard. By spinning up this persistent server, multiple clients can utilize the vector knowledge base endpoints. network.

## Prerequisites

1. Ensure you have built the application (`./bin/idx`).
2. Have a working Ollama endpoint accessible (e.g. `http://localhost:11434`) containing your target embedding model (`embeddinggemma` by default).

## Starting the Server

Execute the `http` subcommand under the generic `serve` branch:

```bash
./bin/idx serve http --dir ./notes --http.address ":8080"
```

### Configuration Options

* `--dir`: The root directory pointing to the `.md` resources you intend to parse (default: `.`)
* `--http.address`: The network binding address for the HTTP listener (default: `:8080`).
* `--ollama.host`: Your active Ollama service (default: `http://localhost:11434`).
* `--ollama.model`: Target embedding generation model.

## Connecting an MCP Client

Depending on your MCP client's supported transport, you have two endpoints available:

### Streamable HTTP (Recommended)

If your client supports the modern Streamable HTTP transport type (such as the default in `mark3labs/mcp-go`), use the following endpoint:

* **Endpoint URL**: `http://localhost:8080/mcp`

### Server-Sent Events (SSE)

For older MCP clients or those strictly requiring the SSE transport standard, use the paired SSE endpoints:

* **SSE Connection URL**: `http://localhost:8080/sse`
* **Message Binding URL**: `http://localhost:8080/message`

Refer to your respective LLM tool client documentation (Claude Desktop, Cursor, generic MCP JS/Python clients) on establishing an active SSE hook using your configured port mapped against `/sse`.

## Interacting via curl

You can also interact with the MCP server directly using `curl` to send JSON-RPC requests.

Because MCP over HTTP relies on Server-Sent Events (SSE) and session identifiers, this requires two terminal windows.

**Terminal 1: Start the SSE Connection**

First, open a connection to the `/sse` endpoint. This will stream events from the server to your client. Use `-N` to prevent curl from buffering the output.

```bash
curl -N http://localhost:8080/sse
```

When you connect, the server will immediately emit an `endpoint` event containing a URL with your unique `sessionId`.

```text
event: endpoint
data: /message?sessionId=12345678-1234-1234-1234-123456789abc
```

**Terminal 2: Send JSON-RPC Messages**

Using the URL provided in the `endpoint` event from Terminal 1, you can now `POST` JSON-RPC payloads to interact with the server.

For example, to list available tools:

```bash
curl -X POST "http://localhost:8080/message?sessionId=YOUR_SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list"
  }'
```

To call the `search_knowledge_base` tool:

```bash
curl -X POST "http://localhost:8080/message?sessionId=YOUR_SESSION_ID" \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "search_knowledge_base",
      "arguments": {
        "query": "your search term here"
      }
    }
  }'
```

You will see the response to your POST request emitted as a `message` event in **Terminal 1**.
