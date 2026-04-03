# idx

Short for "index", latin for "informer".

`idx` is a standalone, purely embedded Go binary that transforms a directory of markdown files into an intelligent Knowledge Base, served securely via the Model Context Protocol (MCP). It features an in-memory vector database powered by `chromem-go`, eliminating the need for complex, hosted vector-search cloud stacks.

## Quick Start

Download or build the CLI from your repository:

```bash
task build
./bin/idx --help
```

Use the tooling to dynamically spin up standard MCP interfaces referencing the `.md` files nested within your target directories, powered by your local Ollama embeddings!

## User Documentation

Comprehensive guides on utilizing `idx` are available in our documentation portal based on the [Diátaxis framework](https://diataxis.fr).

### How-to Guides
* [Serving MCP over Standard IO (stdio)](docs/how-to/serve-mcp-over-stdio.md)
* [Serving MCP over HTTP (SSE)](docs/how-to/serve-mcp-over-http.md)

## Developer Architecture
Please refer to the following root-level documents mapping out our component guidelines:
* Check out `Taskfile.yml` for linting, testing, and generation rules.
