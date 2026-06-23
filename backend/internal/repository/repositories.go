package repository

import (
	"context"
	"time"

	"gorm.io/gorm"
	"ragserver/backend/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) Get(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ListActive(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Where("status = ?", model.StatusActive).Order("id desc").Find(&users).Error
	return users, err
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

type KnowledgeBaseRepository struct {
	db *gorm.DB
}

func NewKnowledgeBaseRepository(db *gorm.DB) *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{db: db}
}

func (r *KnowledgeBaseRepository) Create(ctx context.Context, kb *model.KnowledgeBase) error {
	return r.db.WithContext(ctx).Create(kb).Error
}

func (r *KnowledgeBaseRepository) ListVisible(ctx context.Context, userID uint64) ([]model.KnowledgeBase, error) {
	var items []model.KnowledgeBase
	err := r.db.WithContext(ctx).
		Where("visibility = ? OR owner_user_id = ?", model.VisibilityPublic, userID).
		Order("id desc").
		Find(&items).Error
	return items, err
}

func (r *KnowledgeBaseRepository) Get(ctx context.Context, id uint64) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := r.db.WithContext(ctx).First(&kb, id).Error; err != nil {
		return nil, err
	}
	return &kb, nil
}

func (r *KnowledgeBaseRepository) GetVisible(ctx context.Context, id, userID uint64) (*model.KnowledgeBase, error) {
	var kb model.KnowledgeBase
	if err := r.db.WithContext(ctx).
		Where("id = ? AND (visibility = ? OR owner_user_id = ?)", id, model.VisibilityPublic, userID).
		First(&kb).Error; err != nil {
		return nil, err
	}
	return &kb, nil
}

func (r *KnowledgeBaseRepository) FindVisibleByName(ctx context.Context, userID uint64, name string) ([]model.KnowledgeBase, error) {
	var items []model.KnowledgeBase
	err := r.db.WithContext(ctx).
		Where("name = ? AND (visibility = ? OR owner_user_id = ?)", name, model.VisibilityPublic, userID).
		Find(&items).Error
	return items, err
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

func (r *DocumentRepository) ListByKB(ctx context.Context, kbID uint64, page, pageSize int) ([]model.Document, int64, error) {
	var items []model.Document
	var total int64
	query := r.db.WithContext(ctx).Where("kb_id = ?", kbID)
	if err := query.Model(&model.Document{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *DocumentRepository) Update(ctx context.Context, doc *model.Document) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *DocumentRepository) FailStaleIndexing(ctx context.Context, olderThan time.Time, message string) error {
	return r.db.WithContext(ctx).
		Model(&model.Document{}).
		Where("index_status = ? AND updated_at < ?", model.IndexIndexing, olderThan).
		Updates(map[string]any{
			"index_status": model.IndexFailed,
			"index_error":  message,
		}).Error
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

func (r *IngestionJobRepository) FailStaleRunning(ctx context.Context, olderThan time.Time, message string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.IngestionJob{}).
		Where("status = ? AND updated_at < ?", model.JobRunning, olderThan).
		Updates(map[string]any{
			"status":        model.JobFailed,
			"error_message": message,
			"finished_at":   now,
		}).Error
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

func (r *APIKeyRepository) ListByCreator(ctx context.Context, creatorID uint64, page, pageSize int) ([]model.APIKey, int64, error) {
	var items []model.APIKey
	var total int64
	query := r.db.WithContext(ctx).Where("created_by_user_id = ?", creatorID)
	if err := query.Model(&model.APIKey{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *APIKeyRepository) ExistsByCreatorAndBoundUser(ctx context.Context, creatorID, boundUserID uint64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.APIKey{}).
		Where("created_by_user_id = ? AND bound_user_id = ?", creatorID, boundUserID).
		Count(&count).Error
	return count > 0, err
}

func (r *APIKeyRepository) GetByCreator(ctx context.Context, id, creatorID uint64) (*model.APIKey, error) {
	var key model.APIKey
	if err := r.db.WithContext(ctx).Where("id = ? AND created_by_user_id = ?", id, creatorID).First(&key).Error; err != nil {
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
