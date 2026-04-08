package domain

// Document represents a searchable snippet of knowledge in the system.
type Document struct {
	ID       string
	Content  string
	Metadata map[string]string
}
