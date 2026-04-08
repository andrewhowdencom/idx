package cmd

import (
	"log/slog"

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
		defaultIndex := viper.GetString("default-index")
		ollamaHost := viper.GetString("ollama.host")
		ollamaModel := viper.GetString("ollama.model")

		slog.Debug("starting serve stdio", 
			"dirs", dirs,
			"defaultIndex", defaultIndex,
			"ollamaHost", ollamaHost, 
			"ollamaModel", ollamaModel,
		)

		app, err := di.InitializeServer(cmd.Context(), di.AppConfig{
			Dirs:         dirs,
			DefaultIndex: defaultIndex,
			OllamaHost:   ollamaHost,
			OllamaModel:  ollamaModel,
			MaxChunkSize: 1000,
			Concurrency:  5,
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
		defaultIndex := viper.GetString("default-index")
		ollamaHost := viper.GetString("ollama.host")
		ollamaModel := viper.GetString("ollama.model")
		httpAddr := viper.GetString("http.address")

		slog.Debug("starting serve http", 
			"dirs", dirs,
			"defaultIndex", defaultIndex,
			"ollamaHost", ollamaHost, 
			"ollamaModel", ollamaModel,
			"httpAddr", httpAddr,
		)

		app, err := di.InitializeServer(cmd.Context(), di.AppConfig{
			Dirs:         dirs,
			DefaultIndex: defaultIndex,
			OllamaHost:   ollamaHost,
			OllamaModel:  ollamaModel,
			MaxChunkSize: 1000,
			Concurrency:  5,
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
	serveCmd.PersistentFlags().String("default-index", "default", "The name of the default index to query when none is specified")
	serveCmd.PersistentFlags().String("ollama.host", "http://localhost:11434", "Ollama API endpoint")
	serveCmd.PersistentFlags().String("ollama.model", "embeddinggemma", "Ollama embedding model to use")
	
	serveHttpCmd.Flags().String("http.address", ":8080", "Network address to bind the HTTP server to")

	if err := viper.BindPFlags(serveCmd.PersistentFlags()); err != nil {
		panic(err)
	}
	if err := viper.BindPFlags(serveHttpCmd.Flags()); err != nil {
		panic(err)
	}
}
