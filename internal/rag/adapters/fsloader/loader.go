package fsloader

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewhowdencom/idx/internal/rag/domain"
	"github.com/andrewhowdencom/idx/internal/rag/ports"
)

type fsLoader struct {
	maxChunkSize int
}

// New returns a DocumentLoader that reads markdown files from the local filesystem.
func New(maxChunkSize int) ports.DocumentLoader {
	return &fsLoader{
		maxChunkSize: maxChunkSize,
	}
}

func (l *fsLoader) Load(ctx context.Context, location string) ([]domain.Document, error) {
	var docs []domain.Document

	err := filepath.WalkDir(location, func(path string, d os.DirEntry, err error) error {
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

		chunks := chunkString(string(content), l.maxChunkSize)

		for i, chunk := range chunks {
			if strings.TrimSpace(chunk) == "" {
				continue
			}

			docID := fmt.Sprintf("%s-chunk-%d", path, i)
			docs = append(docs, domain.Document{
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
		return nil, fmt.Errorf("failed walking directory %q: %w", location, err)
	}

	return docs, nil
}

func chunkString(text string, maxSize int) []string {
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
