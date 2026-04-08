//go:build wireinject
// +build wireinject

package di

import (
	"context"

	"github.com/andrewhowdencom/idx/internal/rag/adapters/chromemstore"
	"github.com/andrewhowdencom/idx/internal/rag/adapters/fsloader"
	"github.com/andrewhowdencom/idx/internal/rag/adapters/mcpserver"
	"github.com/andrewhowdencom/idx/internal/rag/service"
	"github.com/google/wire"
)

// AppConfig consolidates all configuration required to bootstrap the RAG application.
type AppConfig struct {
	Dirs         map[string]string
	DefaultIndex string
	OllamaHost   string
	OllamaModel  string
	MaxChunkSize int
	Concurrency  int
}

// Map configuration
func provideChromemConfig(cfg AppConfig) chromemstore.Config {
	return chromemstore.Config{
		OllamaHost:  cfg.OllamaHost,
		OllamaModel: cfg.OllamaModel,
		Concurrency: cfg.Concurrency,
	}
}

func provideMCPConfig(cfg AppConfig) mcpserver.Config {
	return mcpserver.Config{
		DefaultIndex: cfg.DefaultIndex,
	}
}

func provideMaxChunkSize(cfg AppConfig) int {
    return cfg.MaxChunkSize
}

// Application bundles the initialized components for use by the CLI.
type Application struct {
	Adapter *mcpserver.MCPAdapter
	Knowledge   service.KnowledgeBase
}

func provideApplication(adapter *mcpserver.MCPAdapter, kb service.KnowledgeBase) *Application {
	return &Application{
		Adapter:   adapter,
		Knowledge: kb,
	}
}

// InitializeServer sets up the entire application using Google Wire.
func InitializeServer(ctx context.Context, cfg AppConfig) (*Application, error) {
	wire.Build(
		provideChromemConfig,
		chromemstore.New,

		provideMaxChunkSize,
		fsloader.New,

		service.New,

		provideMCPConfig,
		mcpserver.New,

		provideApplication,
	)
	return nil, nil
}


