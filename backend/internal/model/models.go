package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	StatusActive   = "active"
	StatusDisabled = "disabled"

	IndexPending  = "pending"
	IndexIndexing = "indexing"
	IndexIndexed  = "indexed"
	IndexFailed   = "failed"

	JobUploadIndex = "upload_index"
	JobReindex     = "reindex"
	JobDeleteIndex = "delete_index"

	JobPending   = "pending"
	JobRunning   = "running"
	JobSucceeded = "succeeded"
	JobFailed    = "failed"
)

type KnowledgeBase struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:128;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"size:32;not null;default:active" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Document struct {
	ID               uint64         `gorm:"primaryKey" json:"id"`
	KBID             uint64         `gorm:"index;not null" json:"kb_id"`
	Filename         string         `gorm:"size:255;not null" json:"filename"`
	OriginalFilename string         `gorm:"size:255;not null" json:"original_filename"`
	FileExt          string         `gorm:"size:32;not null" json:"file_ext"`
	MimeType         string         `gorm:"size:128;not null" json:"mime_type"`
	FileSize         int64          `gorm:"not null" json:"file_size"`
	FileHash         string         `gorm:"size:128;index;not null" json:"file_hash"`
	StoragePath      string         `gorm:"type:text;not null" json:"storage_path"`
	IndexStatus      string         `gorm:"size:32;not null;default:pending" json:"index_status"`
	IndexError       string         `gorm:"type:text" json:"index_error"`
	ChunkCount       int            `gorm:"not null;default:0" json:"chunk_count"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

type DocumentChunk struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	KBID         uint64         `gorm:"index;not null" json:"kb_id"`
	DocumentID   uint64         `gorm:"index;not null" json:"document_id"`
	ChunkIndex   int            `gorm:"not null" json:"chunk_index"`
	Content      string         `gorm:"type:mediumtext;not null" json:"content"`
	ContentHash  string         `gorm:"size:128;not null" json:"content_hash"`
	TokenCount   int            `gorm:"not null;default:0" json:"token_count"`
	RedisKey     string         `gorm:"size:255;uniqueIndex;not null" json:"redis_key"`
	MetadataJSON datatypes.JSON `gorm:"type:json" json:"metadata_json"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type IngestionJob struct {
	ID           uint64     `gorm:"primaryKey" json:"id"`
	KBID         uint64     `gorm:"index;not null" json:"kb_id"`
	DocumentID   uint64     `gorm:"index;not null" json:"document_id"`
	JobType      string     `gorm:"size:32;not null" json:"job_type"`
	Status       string     `gorm:"size:32;index;not null;default:pending" json:"status"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	StartedAt    *time.Time `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type APIKey struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:128;not null" json:"name"`
	KeyHash      string         `gorm:"size:128;uniqueIndex;not null" json:"-"`
	EncryptedKey string         `gorm:"type:text;not null" json:"-"`
	Status       string         `gorm:"size:32;not null;default:active" json:"status"`
	LastUsedAt   *time.Time     `json:"last_used_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
