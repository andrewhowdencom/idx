package mcpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	stdlibhttp "github.com/andrewhowdencom/stdlib/http"
	"github.com/andrewhowdencom/idx/internal/rag/service"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

// MCPAdapter exposes the RAG service to MCP clients.
type MCPAdapter struct {
	mcpServer    *server.MCPServer
	kb           service.KnowledgeBase
	defaultIndex string
}

// Config configures the MCP Adapter.
type Config struct {
	DefaultIndex string
}

// New creates a new MCP adapter bridging `mcp-go` with the core KnowledgeBase.
func New(kb service.KnowledgeBase, cfg Config) *MCPAdapter {
	mcpServer := server.NewMCPServer("idx-rag", "1.0.0", server.WithLogging())
	adapter := &MCPAdapter{
		mcpServer:    mcpServer,
		kb:           kb,
		defaultIndex: cfg.DefaultIndex,
	}
	adapter.registerTools()
	return adapter
}

func (a *MCPAdapter) registerTools() {
	indexNames := a.kb.GetIndexNames()
	description := "Search the indexed markdown knowledge base for relevant context given an inquiry. Always use this to fetch context before answering questions about the repository documentation."
	if len(indexNames) > 0 {
		description += fmt.Sprintf(" Available indexes: %s.", strings.Join(indexNames, ", "))
	}
	if a.defaultIndex != "" {
		description += fmt.Sprintf(" Default index: %s.", a.defaultIndex)
	}

	searchTool := mcp.NewTool("search_knowledge_base",
		mcp.WithDescription(description),
		mcp.WithString("query", mcp.Required(), mcp.Description("The semantic search query string")),
		mcp.WithString("index", mcp.Description("The name of the index to search in. If omitted, uses the default index.")),
	)

	a.mcpServer.AddTool(searchTool, a.handleSearch)
}

func (a *MCPAdapter) handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return nil, fmt.Errorf("query argument missing or invalid: %w", err)
	}

	targetIndex := request.GetString("index", a.defaultIndex)
	if targetIndex == "" {
		targetIndex = a.defaultIndex
	}

	slog.Info("searching knowledge base", "query", query, "index", targetIndex)

	result, err := a.kb.Search(ctx, targetIndex, query, 5)
	if err != nil {
		slog.Error("failed to orchestrate search", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("failed to query database: %v", err)), nil
	}

	return mcp.NewToolResultText(result), nil
}

// ServeStdio starts the MCP server over standard input and output.
func (a *MCPAdapter) ServeStdio() error {
	slog.Info("starting MCP stdio server")
	return server.ServeStdio(a.mcpServer)
}

// ServeHTTPConfig starts the MCP server using SSE/HTTP mappings.
func (a *MCPAdapter) ServeHTTPConfig(ctx context.Context, httpAddr string) error {
	slog.Info("bootstrapping otel tracer provider")
	exporter, err := stdouttrace.New()
	if err != nil {
		return fmt.Errorf("failed to create otel exporter: %w", err)
	}
	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(ctx) }()

	sseServer := server.NewSSEServer(a.mcpServer)
	streamableServer := server.NewStreamableHTTPServer(a.mcpServer, server.WithEndpointPath("/mcp"))

	mux := http.NewServeMux()
	mux.Handle("/sse", sseServer.SSEHandler())
	mux.Handle("/message", sseServer.MessageHandler())
	mux.Handle("/mcp", streamableServer)

	slog.Info("starting stdlib http sse server", "addr", httpAddr)
	srv, err := stdlibhttp.NewServer(httpAddr, mux, stdlibhttp.WithServerTracerProvider(tp))
	if err != nil {
		return fmt.Errorf("failed to create stdlib http server: %w", err)
	}

	return srv.Run()
}
