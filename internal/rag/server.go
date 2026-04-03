package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	stdlibhttp "github.com/andrewhowdencom/stdlib/http"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/philippgille/chromem-go"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
)

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func checkOllamaModelExists(host, model string) error {
	resp, err := http.Get(host + "/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama at %s: %w", host, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama API returned status %d when checking tags", resp.StatusCode)
	}

	var tags ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return fmt.Errorf("failed to parse ollama tags response: %w", err)
	}

	for _, m := range tags.Models {
		if m.Name == model || m.Name == model+":latest" {
			return nil
		}
	}

	return fmt.Errorf("model %q not found. Please run 'ollama pull %s' or specify an available model", model, model)
}

// InitMCP encapsulates parsing and index logic shared across protocol providers
func InitMCP(ctx context.Context, dir, ollamaHost, ollamaModel string) (*server.MCPServer, error) {
	if dir == "" {
		dir = "."
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("invalid directory path %q: %v", dir, err)
	}

	slog.Info("initializing embedded vector database")
	db := chromem.NewDB()
	
	// chromem-go expects the trailing /api for standard Ollama deployments
	if ollamaHost != "" && !strings.HasSuffix(ollamaHost, "/api") && !strings.HasSuffix(ollamaHost, "/api/") {
		ollamaHost = strings.TrimRight(ollamaHost, "/") + "/api"
	}
	
	if err := checkOllamaModelExists(ollamaHost, ollamaModel); err != nil {
		return nil, fmt.Errorf("ollama preflight check failed: %w", err)
	}
	
	embedFunc := chromem.NewEmbeddingFuncOllama(ollamaModel, ollamaHost)
	collection, err := db.CreateCollection("knowledge_base", nil, embedFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	slog.Info("indexing markdown files", "dir", dir)
	if err := indexDirectory(ctx, collection, dir); err != nil {
		return nil, fmt.Errorf("failed to index directory: %w", err)
	}

	mcpServer := server.NewMCPServer("idx-rag", "1.0.0", server.WithLogging())

	searchTool := mcp.NewTool("search_knowledge_base",
		mcp.WithDescription("Search the indexed markdown knowledge base for relevant context given an inquiry. Always use this to fetch context before answering questions about the repository documentation."),
		mcp.WithString("query", mcp.Required(), mcp.Description("The semantic search query string")),
	)

	mcpServer.AddTool(searchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return nil, fmt.Errorf("query argument missing or invalid: %w", err)
		}

		slog.Info("searching knowledge base", "query", query)

		res, err := collection.Query(context.Background(), query, 5, nil, nil)
		if err != nil {
			slog.Error("failed to query collection", "error", err)
			return mcp.NewToolResultError("failed to query database"), nil
		}

		if len(res) == 0 {
			return mcp.NewToolResultText("No relevant context found in the knowledge base."), nil
		}

		var b strings.Builder
		b.WriteString("Here is the relevant context from the knowledge base:\n\n")
		for i, doc := range res {
			b.WriteString(fmt.Sprintf("---\nDocument: %s\nSimilarity Score: %.3f\nContent:\n%s\n", 
				doc.Metadata["path"], doc.Similarity, doc.Content))
			if i < len(res)-1 {
				b.WriteString("\n")
			}
		}

		return mcp.NewToolResultText(b.String()), nil
	})

	return mcpServer, nil
}

// ServeStdio powers the standard input / standard output protocol binding
func ServeStdio(ctx context.Context, dir, ollamaHost, ollamaModel string) error {
	mcpServer, err := InitMCP(ctx, dir, ollamaHost, ollamaModel)
	if err != nil {
		return err
	}
	slog.Info("starting MCP stdio server")
	return server.ServeStdio(mcpServer)
}

// ServeHTTPConfig powers Server-Sent Events (SSE) RPC standard
func ServeHTTPConfig(ctx context.Context, dir, ollamaHost, ollamaModel, httpAddr string) error {
	mcpServer, err := InitMCP(ctx, dir, ollamaHost, ollamaModel)
	if err != nil {
		return err
	}

	// Bootstrap OTel Base logic
	slog.Info("bootstrapping otel tracer provider")
	exporter, err := stdouttrace.New()
	if err != nil {
		return fmt.Errorf("failed to create otel exporter: %w", err)
	}
	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(ctx) }()

	// Wrap inside SSE implementation
	sseServer := server.NewSSEServer(mcpServer)

	// Combine endpoints using Go's standard multiplexer
	mux := http.NewServeMux()
	mux.Handle("/sse", sseServer.SSEHandler())
	mux.Handle("/message", sseServer.MessageHandler())

	// Implement with custom stdlib wrapper allowing proper config and OTEL traces
	slog.Info("starting stdlib http sse server", "addr", httpAddr)
	srv, err := stdlibhttp.NewServer(httpAddr, mux, stdlibhttp.WithServerTracerProvider(tp))
	if err != nil {
		return fmt.Errorf("failed to create stdlib http server: %w", err)
	}

	return srv.Run()
}

func indexDirectory(ctx context.Context, collection *chromem.Collection, dir string) error {
	var docs []chromem.Document
	
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir // Skip hidden directories like .git
			}
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) != ".md" {
			return nil // Skip non-markdown files
		}

		slog.Debug("reading file", "path", path)
		content, err := os.ReadFile(path)
		if err != nil {
			slog.Warn("failed to read file", "path", path, "error", err)
			return nil
		}

		chunks := simpleChunking(string(content), 1000)
		
		for i, chunk := range chunks {
			if strings.TrimSpace(chunk) == "" {
				continue
			}
			
			docID := fmt.Sprintf("%s-chunk-%d", path, i)
			docs = append(docs, chromem.Document{
				ID:      docID,
				Content: chunk,
				Metadata: map[string]string{
					"path": path,
				},
			})
		}
		return nil
	})

	if err != nil {
		return err
	}

	if len(docs) > 0 {
		slog.Info("embedding documents", "count", len(docs))
		err = collection.AddDocuments(ctx, docs, 100) 
		if err != nil {
			return err
		}
	} else {
		slog.Warn("no markdown files found to embed")
	}

	return nil
}

func simpleChunking(text string, maxSize int) []string {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	var currentChunk strings.Builder

	for _, p := range paragraphs {
		if currentChunk.Len()+len(p) > maxSize && currentChunk.Len() > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
		}
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
		}
		currentChunk.WriteString(p)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}
