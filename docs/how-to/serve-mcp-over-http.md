# How to serve MCP over HTTP (SSE)

`idx` provides a Server-Sent Events (SSE) server for remote or networked clients compatible with the Model Context Protocol (MCP) HTTP standard. By spinning up this persistent server, multiple clients can utilize the vector knowledge base endpoints. network.

## Prerequisites

1. Ensure you have built the application (`./bin/idx`).
2. Have a working Ollama endpoint accessible (e.g. `http://localhost:11434`) containing your target embedding model (`mxbai-embed-large` by default).

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

When your server is running, the protocol exposes the endpoints required to interface with standard MCP clients:

* **SSE Connection URL**: `http://localhost:8080/sse`
* **Message Binding URL**: `http://localhost:8080/message`

Refer to your respective LLM tool client documentation (Claude Desktop, Cursor, generic MCP JS/Python clients) on establishing an active SSE hook using your configured port mapped against `/sse`.
