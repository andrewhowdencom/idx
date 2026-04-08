package fsloader

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoader_Load(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fsloader-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create test files
	md1Content := "First paragraph.\n\nSecond paragraph."
	if err := os.WriteFile(filepath.Join(tempDir, "test1.md"), []byte(md1Content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create sub directory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "test2.md"), []byte("Subdir paragraph."), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create a hidden directory that should be ignored
	hiddenDir := filepath.Join(tempDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatalf("failed to create hidden dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "test3.md"), []byte("Hidden paragraph."), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create a non-md file that should be ignored
	if err := os.WriteFile(filepath.Join(tempDir, "test4.txt"), []byte("Text file."), 0644); err != nil {
		t.Fatalf("failed to write text file: %v", err)
	}

	loader := New(25) // Small chunk size to split test1.md

	docs, err := loader.Load(context.Background(), tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}

	// We expect 2 chunks from test1.md and 1 from test2.md
	// The exact IDs depend on walking order, so we count them.
	test1Count := 0
	test2Count := 0

	for _, doc := range docs {
		if filepath.Base(doc.Metadata["path"]) == "test1.md" {
			test1Count++
		} else if filepath.Base(doc.Metadata["path"]) == "test2.md" {
			test2Count++
		} else {
			t.Errorf("unexpected file in documents: %s", doc.Metadata["path"])
		}
	}

	if test1Count != 2 {
		t.Errorf("expected 2 chunks for test1.md, got %d", test1Count)
	}
	if test2Count != 1 {
		t.Errorf("expected 1 chunk for test2.md, got %d", test2Count)
	}
}

func Test_chunkString(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxSize  int
		expected int // number of chunks we expect
	}{
		{
			name:     "single paragraph below max size",
			text:     "Hello world.",
			maxSize:  20,
			expected: 1,
		},
		{
			name:     "two paragraphs split correctly",
			text:     "Hello world.\n\nThis is a new line.",
			maxSize:  20,
			expected: 2, // "Hello world." (12) + "This is a new line." (19)
		},
		{
			name:     "paragraphs combined if below max size",
			text:     "A.\n\nB.",
			maxSize:  20,
			expected: 1, // "A.\n\nB." is length 6, fits in 20.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunks := chunkString(tt.text, tt.maxSize)
			if len(chunks) != tt.expected {
				t.Errorf("chunkString() got %d chunks, want %d", len(chunks), tt.expected)
			}
		})
	}
}
