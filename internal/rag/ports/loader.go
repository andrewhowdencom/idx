package ports

import (
	"context"

	"github.com/andrewhowdencom/idx/internal/rag/domain"
)

// DocumentLoader defines how to load and chunk source files into Documents.
type DocumentLoader interface {
	// Load parses an abstract location (like a directory or URI) and returns a slice of parsed domain documents.
	Load(ctx context.Context, location string) ([]domain.Document, error)
}
