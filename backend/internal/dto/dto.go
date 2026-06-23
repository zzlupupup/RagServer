package dto

import "time"

type ErrorResponse struct {
	Error string `json:"error"`
}

type PaginationResponse[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type UserResponse struct {
	ID          uint64     `json:"id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Role        string     `json:"role"`
	Status      string     `json:"status,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Role        string `json:"role" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
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
	ID                   uint64    `json:"id"`
	OwnerUserID          uint64    `json:"owner_user_id"`
	OwnerUserDisplayName string    `json:"owner_user_display_name,omitempty"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	Visibility           string    `json:"visibility"`
	Status               string    `json:"status"`
	DocumentCount        int64     `json:"document_count"`
	CanManage            bool      `json:"can_manage"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type KnowledgeBaseListResponse struct {
	Items []KnowledgeBaseResponse `json:"items"`
}

type DocumentResponse struct {
	ID                        uint64    `json:"id"`
	KBID                      uint64    `json:"kb_id"`
	UploadedByUserID          uint64    `json:"uploaded_by_user_id"`
	UploadedByUserDisplayName string    `json:"uploaded_by_user_display_name,omitempty"`
	Filename                  string    `json:"filename"`
	OriginalFilename          string    `json:"original_filename"`
	FileExt                   string    `json:"file_ext"`
	MimeType                  string    `json:"mime_type"`
	FileSize                  int64     `json:"file_size"`
	FileHash                  string    `json:"file_hash"`
	StoragePath               string    `json:"storage_path"`
	IndexStatus               string    `json:"index_status"`
	IndexError                string    `json:"index_error,omitempty"`
	ChunkCount                int       `json:"chunk_count"`
	CanDelete                 bool      `json:"can_delete"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type UploadFileResponse struct {
	KBName      string `json:"kb_name,omitempty"`
	KBID        uint64 `json:"kb_id,omitempty"`
	DocumentID  uint64 `json:"document_id"`
	IndexStatus string `json:"index_status"`
	ChunkCount  int    `json:"chunk_count"`
}

type APIKeyCreateRequest struct {
	Name        string `json:"name"`
	BoundUserID uint64 `json:"bound_user_id" binding:"required"`
}

type APIKeyResponse struct {
	ID                   uint64     `json:"id"`
	CreatedByUserID      uint64     `json:"created_by_user_id"`
	BoundUserID          uint64     `json:"bound_user_id"`
	BoundUserEmail       string     `json:"bound_user_email,omitempty"`
	BoundUserDisplayName string     `json:"bound_user_display_name,omitempty"`
	BoundUserRole        string     `json:"bound_user_role,omitempty"`
	Name                 string     `json:"name"`
	Status               string     `json:"status"`
	LastUsedAt           *time.Time `json:"last_used_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	APIKey               string     `json:"api_key,omitempty"`
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
	KBName     string         `json:"kb_name,omitempty"`
	DocumentID uint64         `json:"document_id"`
	ChunkID    uint64         `json:"chunk_id"`
	Filename   string         `json:"filename"`
	Content    string         `json:"content"`
	Score      float64        `json:"score"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type MCPUploadFileRequest struct {
	KBName   string `json:"kb_name" jsonschema:"knowledge base name"`
	FilePath string `json:"file_path" jsonschema:"server temporary file path returned by /api/v1/mcp/files/upload"`
}

type MCPSearchRequest struct {
	KBName string `json:"kb_name" jsonschema:"knowledge base name"`
	Query  string `json:"query" jsonschema:"search query"`
	TopK   int    `json:"top_k,omitempty" jsonschema:"number of chunks to return, default 5, max 20"`
}

type MCPListKBRequest struct{}

type TempUploadResponse struct {
	FilePath string `json:"file_path"`
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	FileSize int64  `json:"file_size"`
}
