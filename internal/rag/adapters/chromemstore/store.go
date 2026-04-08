package chromemstore

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andrewhowdencom/idx/internal/rag/domain"
	"github.com/andrewhowdencom/idx/internal/rag/ports"
	"github.com/philippgille/chromem-go"
)

type store struct {
	db          *chromem.DB
	embedFunc   chromem.EmbeddingFunc
	embedHost   string
	embedModel  string
	concurrency int
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// Config holds the configuration for the Chromem vector store.
type Config struct {
	OllamaHost  string
	OllamaModel string
	Concurrency int
	DBPath      string
}

// New creates a new VectorStore backed by chromem-go and Ollama.
func New(cfg Config) (ports.VectorStore, error) {
	// chromem-go expects the trailing /api for standard Ollama deployments
	if cfg.OllamaHost != "" && !strings.HasSuffix(cfg.OllamaHost, "/api") && !strings.HasSuffix(cfg.OllamaHost, "/api/") {
		cfg.OllamaHost = strings.TrimRight(cfg.OllamaHost, "/") + "/api"
	}

	if err := checkOllamaModelExists(cfg.OllamaHost, cfg.OllamaModel); err != nil {
		return nil, fmt.Errorf("ollama preflight check failed: %w", err)
	}

	var db *chromem.DB
	var err error
	if cfg.DBPath != "" {
		slog.Debug("initializing persistent vector database", "path", cfg.DBPath)
		db, err = chromem.NewPersistentDB(cfg.DBPath, true)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize persistent chromem db: %w", err)
		}
	} else {
		slog.Debug("initializing in-memory vector database")
		db = chromem.NewDB()
	}

	embedFunc := chromem.NewEmbeddingFuncOllama(cfg.OllamaModel, cfg.OllamaHost)

	concurrency := cfg.Concurrency
	if concurrency <= 0 {
		concurrency = 5
	}

	return &store{
		db:          db,
		embedFunc:   embedFunc,
		embedHost:   cfg.OllamaHost,
		embedModel:  cfg.OllamaModel,
		concurrency: concurrency,
	}, nil
}

func checkOllamaModelExists(host, model string) error {
	resp, err := http.Get(host + "/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama at %s: %w", host, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

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

func (s *store) CreateCollection(ctx context.Context, name string) error {
	_, err := s.db.CreateCollection(name, nil, s.embedFunc)
	if err != nil {
		return fmt.Errorf("failed to create collection %q: %w", name, err)
	}
	return nil
}

func (s *store) AddDocuments(ctx context.Context, collectionName string, docs []domain.Document) error {
	collection := s.db.GetCollection(collectionName, nil)
	if collection == nil {
		return fmt.Errorf("collection %q does not exist", collectionName)
	}

	if len(docs) == 0 {
		return nil
	}

	slog.Info("embedding documents", "collection", collectionName, "count", len(docs))

	sem := make(chan struct{}, s.concurrency)
	var wg sync.WaitGroup

	for _, domainDoc := range docs {
		wg.Add(1)
		sem <- struct{}{}

		// Convert domain.Document to chromem.Document
		chromemDoc := chromem.Document{
			ID:       domainDoc.ID,
			Content:  domainDoc.Content,
			Metadata: domainDoc.Metadata,
		}

		go func(doc chromem.Document) {
			defer wg.Done()
			defer func() { <-sem }()

			var err error
			maxRetries := 3
			for attempt := 1; attempt <= maxRetries; attempt++ {
				err = collection.AddDocuments(ctx, []chromem.Document{doc}, 1)
				if err == nil {
					return // Success
				}

				slog.Warn("embedding generation failed",
					"document", doc.Metadata["path"],
					"chunk_id", doc.ID,
					"attempt", attempt,
					"max_retries", maxRetries,
					"error", err,
				)

				select {
				case <-ctx.Done():
					return // Context cancelled, exit early
				case <-time.After(time.Duration(attempt*attempt) * time.Second):
					// Wait and retry
				}
			}

			if err != nil {
				slog.Error("skipping document chunk due to repeated embedding failures",
					"document", doc.Metadata["path"],
					"chunk_id", doc.ID,
					"error", err,
				)
			}
		}(chromemDoc)
	}

	wg.Wait()
	return nil
}

func (s *store) Query(ctx context.Context, collectionName string, query string, n int) ([]ports.DocumentScore, error) {
	collection := s.db.GetCollection(collectionName, nil)
	if collection == nil {
		return nil, fmt.Errorf("collection %q does not exist", collectionName)
	}

	res, err := collection.Query(ctx, query, n, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query chromem collection: %w", err)
	}

	var results []ports.DocumentScore
	for _, doc := range res {
		results = append(results, ports.DocumentScore{
			Document: domain.Document{
				ID:       doc.ID,
				Content:  doc.Content,
				Metadata: doc.Metadata,
			},
			Score: doc.Similarity,
		})
	}

	return results, nil
}
