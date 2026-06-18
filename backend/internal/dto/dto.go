package dto

import "time"

type ErrorResponse struct {
	Error string `json:"error"`
}

type KnowledgeBaseCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type KnowledgeBaseUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type KnowledgeBaseResponse struct {
	ID            uint64    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	DocumentCount int64     `json:"document_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type KnowledgeBaseListResponse struct {
	Items []KnowledgeBaseResponse `json:"items"`
}

type DocumentResponse struct {
	ID               uint64    `json:"id"`
	KBID             uint64    `json:"kb_id"`
	Filename         string    `json:"filename"`
	OriginalFilename string    `json:"original_filename"`
	FileExt          string    `json:"file_ext"`
	MimeType         string    `json:"mime_type"`
	FileSize         int64     `json:"file_size"`
	FileHash         string    `json:"file_hash"`
	StoragePath      string    `json:"storage_path"`
	IndexStatus      string    `json:"index_status"`
	IndexError       string    `json:"index_error,omitempty"`
	ChunkCount       int       `json:"chunk_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type UploadFileResponse struct {
	KBID        uint64 `json:"kb_id"`
	DocumentID  uint64 `json:"document_id"`
	IndexStatus string `json:"index_status"`
	ChunkCount  int    `json:"chunk_count"`
}

type APIKeyCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

type APIKeyResponse struct {
	ID         uint64     `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	LastUsedAt *time.Time `json:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	APIKey     string     `json:"api_key,omitempty"`
}

type RevealAPIKeyResponse struct {
	APIKey string `json:"api_key"`
}

type SearchRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k"`
}

type SearchResponse struct {
	Summary string       `json:"summary"`
	Items   []SearchItem `json:"items"`
}

type SearchItem struct {
	DocumentID uint64         `json:"document_id"`
	ChunkID    uint64         `json:"chunk_id"`
	Filename   string         `json:"filename"`
	Content    string         `json:"content"`
	Score      float64        `json:"score"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type MCPUploadFileRequest struct {
	KBID          uint64 `json:"kb_id" jsonschema:"knowledge base id"`
	Filename      string `json:"filename" jsonschema:"file name"`
	MimeType      string `json:"mime_type" jsonschema:"mime type"`
	ContentBase64 string `json:"content_base64" jsonschema:"base64 encoded file content"`
}

type MCPSearchRequest struct {
	KBID  uint64 `json:"kb_id" jsonschema:"knowledge base id"`
	Query string `json:"query" jsonschema:"search query"`
	TopK  int    `json:"top_k,omitempty" jsonschema:"number of chunks to return, default 5, max 20"`
}

type MCPListKBRequest struct{}
