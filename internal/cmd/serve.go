package cmd

import (
	"log/slog"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/andrewhowdencom/idx/internal/rag/di"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the RAG server (see subcommands for protocols)",
}

var serveStdioCmd = &cobra.Command{
	Use:   "stdio",
	Short: "Starts the MCP server over Standard IO",
	RunE: func(cmd *cobra.Command, args []string) error {
		dirs := viper.GetStringMapString("dir")
		defaultIndex := viper.GetString("index.default")
		ollamaHost := viper.GetString("ollama.host")
		ollamaModel := viper.GetString("ollama.model")
		dbPath := viper.GetString("db.path")
		concurrency := viper.GetInt("concurrency")
		chunkSize := viper.GetInt("chunk-size")

		slog.Debug("starting serve stdio", 
			"dirs", dirs,
			"defaultIndex", defaultIndex,
			"ollamaHost", ollamaHost, 
			"ollamaModel", ollamaModel,
			"dbPath", dbPath,
			"concurrency", concurrency,
			"chunkSize", chunkSize,
		)

		app, err := di.InitializeServer(cmd.Context(), di.AppConfig{
			Dirs:         dirs,
			DefaultIndex: defaultIndex,
			OllamaHost:   ollamaHost,
			OllamaModel:  ollamaModel,
			MaxChunkSize: chunkSize,
			Concurrency:  concurrency,
			DBPath:       dbPath,
		})
		if err != nil {
			return err
		}

		if err := app.Knowledge.InitIndex(cmd.Context(), dirs); err != nil {
			return err
		}

		return app.Adapter.ServeStdio()
	},
}

var serveHttpCmd = &cobra.Command{
	Use:   "http",
	Short: "Starts the MCP server over HTTP/SSE",
	RunE: func(cmd *cobra.Command, args []string) error {
		dirs := viper.GetStringMapString("dir")
		defaultIndex := viper.GetString("index.default")
		ollamaHost := viper.GetString("ollama.host")
		ollamaModel := viper.GetString("ollama.model")
		dbPath := viper.GetString("db.path")
		concurrency := viper.GetInt("concurrency")
		chunkSize := viper.GetInt("chunk-size")
		httpAddr := viper.GetString("http.address")

		slog.Debug("starting serve http", 
			"dirs", dirs,
			"defaultIndex", defaultIndex,
			"ollamaHost", ollamaHost, 
			"ollamaModel", ollamaModel,
			"dbPath", dbPath,
			"concurrency", concurrency,
			"chunkSize", chunkSize,
			"httpAddr", httpAddr,
		)

		app, err := di.InitializeServer(cmd.Context(), di.AppConfig{
			Dirs:         dirs,
			DefaultIndex: defaultIndex,
			OllamaHost:   ollamaHost,
			OllamaModel:  ollamaModel,
			MaxChunkSize: chunkSize,
			Concurrency:  concurrency,
			DBPath:       dbPath,
		})
		if err != nil {
			return err
		}

		if err := app.Knowledge.InitIndex(cmd.Context(), dirs); err != nil {
			return err
		}

		return app.Adapter.ServeHTTPConfig(cmd.Context(), httpAddr)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.AddCommand(serveStdioCmd)
	serveCmd.AddCommand(serveHttpCmd)

	serveCmd.PersistentFlags().StringToString("dir", map[string]string{"default": "."}, "Directories containing markdown files to index (e.g., --dir name=path)")
	serveCmd.PersistentFlags().String("index.default", "default", "The name of the default index to query when none is specified")
	serveCmd.PersistentFlags().String("ollama.host", "http://localhost:11434", "Ollama API endpoint")
	serveCmd.PersistentFlags().String("ollama.model", "embeddinggemma", "Ollama embedding model to use")
	serveCmd.PersistentFlags().String("db.path", filepath.Join(xdg.DataHome, "idx", "db"), "Path to store the vector database (empty for in-memory)")
	serveCmd.PersistentFlags().Int("concurrency", 5, "Number of concurrent embedding jobs")
	serveCmd.PersistentFlags().Int("chunk-size", 1000, "Maximum size for markdown chunks")
	
	serveHttpCmd.Flags().String("http.address", ":8080", "Network address to bind the HTTP server to")

	if err := viper.BindPFlags(serveCmd.PersistentFlags()); err != nil {
		panic(err)
	}
	if err := viper.BindPFlags(serveHttpCmd.Flags()); err != nil {
		panic(err)
	}
}
