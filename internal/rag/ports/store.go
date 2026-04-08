package ports

import (
	"context"

	"github.com/andrewhowdencom/idx/internal/rag/domain"
)

// DocumentScore represents a matched document with its similarity score.
type DocumentScore struct {
	Document domain.Document
	Score    float32
}

// VectorStore defines the interface for interacting with a vector database.
type VectorStore interface {
	// CreateCollection creates a new index inside the vector store.
	CreateCollection(ctx context.Context, name string) error
	
	// AddDocuments embeds and stores the domain documents into the specified collection.
	AddDocuments(ctx context.Context, collectionName string, docs []domain.Document) error
	
	// Query searches the specified collection for the top n most relevant documents to the query string.
	Query(ctx context.Context, collectionName string, query string, n int) ([]DocumentScore, error)
}
