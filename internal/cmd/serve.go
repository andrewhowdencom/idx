package cmd

import (
	"log/slog"

	"github.com/andrewhowdencom/idx/internal/rag"
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
		dir := viper.GetString("dir")
		ollamaHost := viper.GetString("ollama.host")
		ollamaModel := viper.GetString("ollama.model")

		slog.Debug("starting serve stdio", 
			"dir", dir, 
			"ollamaHost", ollamaHost, 
			"ollamaModel", ollamaModel,
		)

		return rag.ServeStdio(cmd.Context(), dir, ollamaHost, ollamaModel)
	},
}

var serveHttpCmd = &cobra.Command{
	Use:   "http",
	Short: "Starts the MCP server over HTTP/SSE",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := viper.GetString("dir")
		ollamaHost := viper.GetString("ollama.host")
		ollamaModel := viper.GetString("ollama.model")
		httpAddr := viper.GetString("http.address")

		slog.Debug("starting serve http", 
			"dir", dir, 
			"ollamaHost", ollamaHost, 
			"ollamaModel", ollamaModel,
			"httpAddr", httpAddr,
		)

		return rag.ServeHTTPConfig(cmd.Context(), dir, ollamaHost, ollamaModel, httpAddr)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.AddCommand(serveStdioCmd)
	serveCmd.AddCommand(serveHttpCmd)

	serveCmd.PersistentFlags().String("dir", ".", "Directory containing markdown files to index")
	serveCmd.PersistentFlags().String("ollama.host", "http://localhost:11434", "Ollama API endpoint")
	serveCmd.PersistentFlags().String("ollama.model", "mxbai-embed-large", "Ollama embedding model to use")
	
	serveHttpCmd.Flags().String("http.address", ":8080", "Network address to bind the HTTP server to")

	if err := viper.BindPFlags(serveCmd.PersistentFlags()); err != nil {
		panic(err)
	}
	if err := viper.BindPFlags(serveHttpCmd.Flags()); err != nil {
		panic(err)
	}
}
