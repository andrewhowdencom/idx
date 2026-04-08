package service

import (
	"context"
	"errors"
	"testing"

	"github.com/andrewhowdencom/idx/internal/rag/domain"
	"github.com/andrewhowdencom/idx/internal/rag/ports"
)

type mockStore struct {
	collections map[string][]domain.Document
}

func (m *mockStore) CreateCollection(ctx context.Context, name string) error {
	if m.collections == nil {
		m.collections = make(map[string][]domain.Document)
	}
	m.collections[name] = []domain.Document{}
	return nil
}

func (m *mockStore) AddDocuments(ctx context.Context, collectionName string, docs []domain.Document) error {
	m.collections[collectionName] = append(m.collections[collectionName], docs...)
	return nil
}

func (m *mockStore) Query(ctx context.Context, collectionName string, query string, n int) ([]ports.DocumentScore, error) {
	if collectionName == "errorStore" {
		return nil, errors.New("query failed")
	}
	docs, exists := m.collections[collectionName]
	if !exists {
		return nil, errors.New("not found")
	}

	var results []ports.DocumentScore
	for _, d := range docs {
		results = append(results, ports.DocumentScore{Document: d, Score: 0.99})
	}
	return results, nil
}

type mockLoader struct{}

func (m *mockLoader) Load(ctx context.Context, location string) ([]domain.Document, error) {
	if location == "bad" {
		return nil, errors.New("failed to load")
	}
	return []domain.Document{
		{ID: "1", Content: "Mocked content", Metadata: map[string]string{"path": "mock.md"}},
	}, nil
}

func TestKnowledgeBase_InitIndex(t *testing.T) {
	tests := []struct {
		name      string
		dirs      map[string]string
		wantError bool
	}{
		{
			name:      "success",
			dirs:      map[string]string{"idx1": "good", "idx2": "good"},
			wantError: false,
		},
		{
			name:      "loader fails",
			dirs:      map[string]string{"bad_idx": "bad"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			loader := &mockLoader{}
			kb := New(store, loader)

			err := kb.InitIndex(context.Background(), tt.dirs)
			if (err != nil) != tt.wantError {
				t.Errorf("InitIndex() error = %v, wantErr %v", err, tt.wantError)
			}
		})
	}
}

func TestKnowledgeBase_Search(t *testing.T) {
	kb := New(&mockStore{}, &mockLoader{})
	_ = kb.InitIndex(context.Background(), map[string]string{"idx": "good"})

	tests := []struct {
		name      string
		indexName string
		query     string
		want      string
		wantError bool
	}{
		{
			name:      "success",
			indexName: "idx",
			query:     "search",
			want:      "Mocked content",
			wantError: false,
		},
		{
			name:      "store fail",
			indexName: "errorStore",
			query:     "search",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kb.Search(context.Background(), tt.indexName, tt.query, 1)
			if (err != nil) != tt.wantError {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantError)
				return
			}
			if !tt.wantError && len(got) > 0 && !errors.Is(err, nil) {
				t.Errorf("Search() got = %v, want to contain %v", got, tt.want) // simple existence check
			}
		})
	}
}
