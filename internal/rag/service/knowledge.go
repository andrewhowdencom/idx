package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/andrewhowdencom/idx/internal/rag/ports"
)

// KnowledgeBase provides the core business logic for indexing and searching.
type KnowledgeBase interface {
	InitIndex(ctx context.Context, dirs map[string]string) error
	Search(ctx context.Context, indexName, query string, limit int) (string, error)
	GetIndexNames() []string
}

type knowledgeBase struct {
	store      ports.VectorStore
	loader     ports.DocumentLoader
	indexNames []string
}

// New returns a new KnowledgeBase service.
func New(store ports.VectorStore, loader ports.DocumentLoader) KnowledgeBase {
	return &knowledgeBase{
		store:  store,
		loader: loader,
	}
}

func (kb *knowledgeBase) InitIndex(ctx context.Context, dirs map[string]string) error {
	var indexNames []string

	for name, dirPath := range dirs {
		if dirPath == "" {
			dirPath = "."
		}

		slog.Info("initializing collection", "name", name)
		if err := kb.store.CreateCollection(ctx, name); err != nil {
			return fmt.Errorf("failed to create collection %q: %w", name, err)
		}

		docs, err := kb.loader.Load(ctx, dirPath)
		if err != nil {
			return fmt.Errorf("failed to load documents for index %q from %q: %w", name, dirPath, err)
		}

		if len(docs) > 0 {
			if err := kb.store.AddDocuments(ctx, name, docs); err != nil {
				return fmt.Errorf("failed to add documents for index %q: %w", name, err)
			}
		} else {
			slog.Warn("no Markdown files found to embed", "index", name, "path", dirPath)
		}

		indexNames = append(indexNames, name)
	}

	kb.indexNames = indexNames
	return nil
}

func (kb *knowledgeBase) Search(ctx context.Context, indexName, query string, limit int) (string, error) {
	docs, err := kb.store.Query(ctx, indexName, query, limit)
	if err != nil {
		return "", fmt.Errorf("failed to query store for index %q: %w", indexName, err)
	}

	if len(docs) == 0 {
		return "No relevant context found in the knowledge base.", nil
	}

	var b strings.Builder
	b.WriteString("Here is the relevant context from the knowledge base:\n\n")
	for i, ds := range docs {
		b.WriteString(fmt.Sprintf("---\nDocument: %s\nSimilarity Score: %.3f\nContent:\n%s\n",
			ds.Document.Metadata["path"], ds.Score, ds.Document.Content))
		if i < len(docs)-1 {
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

func (kb *knowledgeBase) GetIndexNames() []string {
	return kb.indexNames
}
