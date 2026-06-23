package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/rag"
	"ragserver/backend/internal/rag/parser"
	"ragserver/backend/internal/repository"
	appstorage "ragserver/backend/internal/storage"
)

type DocumentService struct {
	kbRepo       *repository.KnowledgeBaseRepository
	docRepo      *repository.DocumentRepository
	chunkRepo    *repository.ChunkRepository
	jobRepo      *repository.IngestionJobRepository
	fileStore    *appstorage.FileStore
	pipeline     *rag.Pipeline
	users        *repository.UserRepository
	indexTimeout time.Duration
}

func NewDocumentService(
	kbRepo *repository.KnowledgeBaseRepository,
	docRepo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	jobRepo *repository.IngestionJobRepository,
	fileStore *appstorage.FileStore,
	pipeline *rag.Pipeline,
	users *repository.UserRepository,
	indexTimeoutSeconds int,
) *DocumentService {
	if indexTimeoutSeconds <= 0 {
		indexTimeoutSeconds = 300
	}
	return &DocumentService{
		kbRepo:       kbRepo,
		docRepo:      docRepo,
		chunkRepo:    chunkRepo,
		jobRepo:      jobRepo,
		fileStore:    fileStore,
		pipeline:     pipeline,
		users:        users,
		indexTimeout: time.Duration(indexTimeoutSeconds) * time.Second,
	}
}

func (s *DocumentService) UploadMultipart(ctx context.Context, user *model.User, kbID uint64, file *multipart.FileHeader) (*dto.DocumentResponse, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return s.UploadBytes(ctx, user, kbID, file.Filename, file.Header.Get("Content-Type"), data)
}

func (s *DocumentService) UploadTempFile(ctx context.Context, user *model.User, kbID uint64, tempPath string) (*dto.DocumentResponse, error) {
	filename, data, err := s.fileStore.ReadTemp(tempPath)
	if err != nil {
		return nil, err
	}
	resp, err := s.UploadBytes(ctx, user, kbID, filename, "", data)
	if err != nil {
		return nil, err
	}
	_ = s.fileStore.DeleteTemp(tempPath)
	return resp, nil
}

func (s *DocumentService) UploadBytes(ctx context.Context, user *model.User, kbID uint64, filename, mimeType string, data []byte) (*dto.DocumentResponse, error) {
	if _, err := s.kbRepo.GetVisible(ctx, kbID, user.ID); err != nil {
		return nil, fmt.Errorf("knowledge base not visible")
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	if !parser.SupportedExtension(filename) {
		return nil, fmt.Errorf("unsupported file type: %s", filepath.Ext(filename))
	}
	doc := &model.Document{
		KBID:             kbID,
		UploadedByUserID: user.ID,
		Filename:         safeStoredName(filename),
		OriginalFilename: filename,
		FileExt:          strings.ToLower(filepath.Ext(filename)),
		MimeType:         mimeType,
		FileSize:         int64(len(data)),
		IndexStatus:      model.IndexPending,
	}
	if err := s.docRepo.Create(ctx, doc); err != nil {
		return nil, err
	}
	path, fileHash, err := s.fileStore.Save(kbID, doc.ID, filename, data)
	if err != nil {
		doc.IndexStatus = model.IndexFailed
		doc.IndexError = err.Error()
		_ = s.docRepo.Update(ctx, doc)
		return nil, err
	}
	doc.StoragePath = path
	doc.FileHash = fileHash
	doc.IndexStatus = model.IndexIndexing
	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}

	job := &model.IngestionJob{
		KBID:            kbID,
		DocumentID:      doc.ID,
		CreatedByUserID: user.ID,
		JobType:         model.JobUploadIndex,
		Status:          model.JobRunning,
	}
	now := time.Now()
	job.StartedAt = &now
	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, s.failDocument(ctx, doc, nil, err)
	}

	// Indexing runs asynchronously so the upload request returns immediately.
	// The frontend polls document status until it reaches a terminal state.
	go s.indexDocument(doc, job, data, user.DisplayName)

	resp := documentToDTO(*doc, true, user.DisplayName)
	return &resp, nil
}

// indexDocument performs chunking + embedding + persistence in the background.
// It uses an independent context so it survives request cancellation.
func (s *DocumentService) indexDocument(doc *model.Document, job *model.IngestionJob, data []byte, uploaderDisplayName string) {
	indexCtx, cancel := context.WithTimeout(context.Background(), s.indexTimeout)
	defer cancel()

	einoDocs, dbChunks, err := s.pipeline.BuildChunks(indexCtx, *doc, data)
	if err == nil {
		err = s.pipeline.Index(indexCtx, einoDocs)
	}
	if err == nil {
		err = s.chunkRepo.ReplaceForDocument(indexCtx, doc.ID, dbChunks)
	}
	if err != nil {
		log.Printf("document_service: indexing failed for doc %d: %v", doc.ID, err)
		s.failDocument(context.Background(), doc, job, err)
		return
	}

	doc.IndexStatus = model.IndexIndexed
	doc.IndexError = ""
	doc.ChunkCount = len(dbChunks)
	if err := s.docRepo.Update(indexCtx, doc); err != nil {
		s.failDocument(context.Background(), doc, job, err)
		return
	}
	_ = s.jobRepo.MarkFinished(indexCtx, job, model.JobSucceeded, "")
}

func (s *DocumentService) failDocument(ctx context.Context, doc *model.Document, job *model.IngestionJob, err error) error {
	doc.IndexStatus = model.IndexFailed
	doc.IndexError = err.Error()
	if uerr := s.docRepo.Update(ctx, doc); uerr != nil {
		log.Printf("document_service: failed to mark doc %d as failed: %v (original error: %v)", doc.ID, uerr, err)
	}
	if job != nil {
		if merr := s.jobRepo.MarkFinished(ctx, job, model.JobFailed, err.Error()); merr != nil {
			log.Printf("document_service: failed to mark job %d as failed: %v (original error: %v)", job.ID, merr, err)
		}
	}
	return err
}

func (s *DocumentService) ListByKB(ctx context.Context, user *model.User, kbID uint64, page, pageSize int) (*dto.PaginationResponse[dto.DocumentResponse], error) {
	kb, err := s.kbRepo.GetVisible(ctx, kbID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not visible")
	}
	page, pageSize = normalizePagination(page, pageSize)
	items, total, err := s.docRepo.ListByKB(ctx, kbID, page, pageSize)
	if err != nil {
		return nil, err
	}
	out := make([]dto.DocumentResponse, 0, len(items))
	for _, item := range items {
		uploaderName := ""
		if u, _ := s.users.Get(ctx, item.UploadedByUserID); u != nil {
			uploaderName = u.DisplayName
		}
		out = append(out, documentToDTO(item, canDeleteDocument(user, kb, item), uploaderName))
	}
	return &dto.PaginationResponse[dto.DocumentResponse]{
		Items:    out,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *DocumentService) GetStatus(ctx context.Context, user *model.User, id uint64) (*dto.DocumentResponse, error) {
	doc, err := s.docRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	kb, err := s.kbRepo.GetVisible(ctx, doc.KBID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not visible")
	}
	uploaderName := ""
	if u, _ := s.users.Get(ctx, doc.UploadedByUserID); u != nil {
		uploaderName = u.DisplayName
	}
	resp := documentToDTO(*doc, canDeleteDocument(user, kb, *doc), uploaderName)
	return &resp, nil
}

func (s *DocumentService) Delete(ctx context.Context, user *model.User, id uint64) error {
	doc, err := s.docRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	kb, err := s.kbRepo.GetVisible(ctx, doc.KBID, user.ID)
	if err != nil {
		return fmt.Errorf("knowledge base not visible")
	}
	if !canDeleteDocument(user, kb, *doc) {
		return fmt.Errorf("not allowed to delete document")
	}
	if err := s.pipeline.DeleteDocumentVectors(ctx, doc.KBID, doc.ID); err != nil {
		return err
	}
	if err := s.chunkRepo.DeleteByDocument(ctx, doc.ID); err != nil {
		return err
	}
	if err := s.fileStore.Delete(doc.StoragePath); err != nil && !strings.Contains(err.Error(), "cannot find") {
		return err
	}
	return s.docRepo.Delete(ctx, id)
}

func documentToDTO(doc model.Document, canDelete bool, uploaderDisplayName string) dto.DocumentResponse {
	return dto.DocumentResponse{
		ID:                        doc.ID,
		KBID:                      doc.KBID,
		UploadedByUserID:          doc.UploadedByUserID,
		UploadedByUserDisplayName: uploaderDisplayName,
		Filename:                  doc.Filename,
		OriginalFilename:          doc.OriginalFilename,
		FileExt:                   doc.FileExt,
		MimeType:                  doc.MimeType,
		FileSize:                  doc.FileSize,
		FileHash:                  doc.FileHash,
		StoragePath:               doc.StoragePath,
		IndexStatus:               doc.IndexStatus,
		IndexError:                doc.IndexError,
		ChunkCount:                doc.ChunkCount,
		CanDelete:                 canDelete,
		CreatedAt:                 doc.CreatedAt,
		UpdatedAt:                 doc.UpdatedAt,
	}
}

func canDeleteDocument(user *model.User, kb *model.KnowledgeBase, doc model.Document) bool {
	return kb.OwnerUserID == user.ID || doc.UploadedByUserID == user.ID
}

func safeStoredName(filename string) string {
	base := filepath.Base(filename)
	if base == "." || base == "" {
		return "upload"
	}
	return base
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
