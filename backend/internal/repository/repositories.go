package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"ragserver/backend/internal/model"
)

type KnowledgeBaseRepository struct {
	db *gorm.DB
}

func NewKnowledgeBaseRepository(db *gorm.DB) *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{db: db}
}

func (r *KnowledgeBaseRepository) Create(ctx context.Context, kb *model.KnowledgeBase) error {
	return r.db.WithContext(ctx).Create(kb).Error
}

func (r *KnowledgeBaseRepository) List(ctx context.Context) ([]model.KnowledgeBase, error) {
	var items []model.KnowledgeBase
	err := r.db.WithContext(ctx).Order("id desc").Find(&items).Error
	return items, err
}

func (r *KnowledgeBaseRepository) Get(ctx context.Context, id uint64) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := r.db.WithContext(ctx).First(&kb, id).Error; err != nil {
		return nil, err
	}
	return &kb, nil
}

func (r *KnowledgeBaseRepository) Update(ctx context.Context, kb *model.KnowledgeBase) error {
	return r.db.WithContext(ctx).Save(kb).Error
}

func (r *KnowledgeBaseRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.KnowledgeBase{}, id).Error
}

func (r *KnowledgeBaseRepository) CountDocuments(ctx context.Context, kbID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Document{}).Where("kb_id = ?", kbID).Count(&count).Error
	return count, err
}

type DocumentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Create(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *DocumentRepository) Get(ctx context.Context, id uint64) (*model.Document, error) {
	var doc model.Document
	if err := r.db.WithContext(ctx).First(&doc, id).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentRepository) ListByKB(ctx context.Context, kbID uint64) ([]model.Document, error) {
	var items []model.Document
	err := r.db.WithContext(ctx).Where("kb_id = ?", kbID).Order("id desc").Find(&items).Error
	return items, err
}

func (r *DocumentRepository) Update(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *DocumentRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Document{}, id).Error
}

type ChunkRepository struct {
	db *gorm.DB
}

func NewChunkRepository(db *gorm.DB) *ChunkRepository {
	return &ChunkRepository{db: db}
}

func (r *ChunkRepository) ReplaceForDocument(ctx context.Context, documentID uint64, chunks []model.DocumentChunk) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("document_id = ?", documentID).Delete(&model.DocumentChunk{}).Error; err != nil {
			return err
		}
		if len(chunks) == 0 {
			return nil
		}
		return tx.Create(&chunks).Error
	})
}

func (r *ChunkRepository) DeleteByDocument(ctx context.Context, documentID uint64) error {
	return r.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&model.DocumentChunk{}).Error
}

func (r *ChunkRepository) GetByRedisKey(ctx context.Context, redisKey string) (*model.DocumentChunk, error) {
	var chunk model.DocumentChunk
	if err := r.db.WithContext(ctx).Where("redis_key = ?", redisKey).First(&chunk).Error; err != nil {
		return nil, err
	}
	return &chunk, nil
}

type IngestionJobRepository struct {
	db *gorm.DB
}

func NewIngestionJobRepository(db *gorm.DB) *IngestionJobRepository {
	return &IngestionJobRepository{db: db}
}

func (r *IngestionJobRepository) Create(ctx context.Context, job *model.IngestionJob) error {
	return r.db.WithContext(ctx).Create(job).Error
}

func (r *IngestionJobRepository) Update(ctx context.Context, job *model.IngestionJob) error {
	return r.db.WithContext(ctx).Save(job).Error
}

func (r *IngestionJobRepository) MarkFinished(ctx context.Context, job *model.IngestionJob, status, message string) error {
	now := time.Now()
	job.Status = status
	job.ErrorMessage = message
	job.FinishedAt = &now
	return r.Update(ctx, job)
}

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(ctx context.Context, key *model.APIKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *APIKeyRepository) List(ctx context.Context) ([]model.APIKey, error) {
	var items []model.APIKey
	err := r.db.WithContext(ctx).Order("id desc").Find(&items).Error
	return items, err
}

func (r *APIKeyRepository) Get(ctx context.Context, id uint64) (*model.APIKey, error) {
	var key model.APIKey
	if err := r.db.WithContext(ctx).First(&key, id).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) GetByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	var key model.APIKey
	if err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) Update(ctx context.Context, key *model.APIKey) error {
	return r.db.WithContext(ctx).Save(key).Error
}

func (r *APIKeyRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.APIKey{}, id).Error
}
