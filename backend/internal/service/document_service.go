package service

import (
	"context"
	"fmt"
	"io"
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
	kbRepo    *repository.KnowledgeBaseRepository
	docRepo   *repository.DocumentRepository
	chunkRepo *repository.ChunkRepository
	jobRepo   *repository.IngestionJobRepository
	fileStore *appstorage.FileStore
	pipeline  *rag.Pipeline
}

func NewDocumentService(
	kbRepo *repository.KnowledgeBaseRepository,
	docRepo *repository.DocumentRepository,
	chunkRepo *repository.ChunkRepository,
	jobRepo *repository.IngestionJobRepository,
	fileStore *appstorage.FileStore,
	pipeline *rag.Pipeline,
) *DocumentService {
	return &DocumentService{
		kbRepo:    kbRepo,
		docRepo:   docRepo,
		chunkRepo: chunkRepo,
		jobRepo:   jobRepo,
		fileStore: fileStore,
		pipeline:  pipeline,
	}
}

func (s *DocumentService) UploadMultipart(ctx context.Context, kbID uint64, file *multipart.FileHeader) (*dto.DocumentResponse, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return s.UploadBytes(ctx, kbID, file.Filename, file.Header.Get("Content-Type"), data)
}

func (s *DocumentService) UploadBytes(ctx context.Context, kbID uint64, filename, mimeType string, data []byte) (*dto.DocumentResponse, error) {
	if _, err := s.kbRepo.Get(ctx, kbID); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	if !parser.SupportedExtension(filename) {
		return nil, fmt.Errorf("unsupported file type: %s", filepath.Ext(filename))
	}
	doc := &model.Document{
		KBID:             kbID,
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
		KBID:       kbID,
		DocumentID: doc.ID,
		JobType:    model.JobUploadIndex,
		Status:     model.JobRunning,
	}
	now := time.Now()
	job.StartedAt = &now
	_ = s.jobRepo.Create(ctx, job)

	einoDocs, dbChunks, err := s.pipeline.BuildChunks(ctx, *doc, data)
	if err == nil {
		err = s.pipeline.Index(ctx, einoDocs)
	}
	if err == nil {
		err = s.chunkRepo.ReplaceForDocument(ctx, doc.ID, dbChunks)
	}
	if err != nil {
		doc.IndexStatus = model.IndexFailed
		doc.IndexError = err.Error()
		_ = s.docRepo.Update(ctx, doc)
		_ = s.jobRepo.MarkFinished(ctx, job, model.JobFailed, err.Error())
		return nil, err
	}

	doc.IndexStatus = model.IndexIndexed
	doc.IndexError = ""
	doc.ChunkCount = len(dbChunks)
	if err := s.docRepo.Update(ctx, doc); err != nil {
		return nil, err
	}
	_ = s.jobRepo.MarkFinished(ctx, job, model.JobSucceeded, "")
	resp := documentToDTO(*doc)
	return &resp, nil
}

func (s *DocumentService) ListByKB(ctx context.Context, kbID uint64) ([]dto.DocumentResponse, error) {
	items, err := s.docRepo.ListByKB(ctx, kbID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.DocumentResponse, 0, len(items))
	for _, item := range items {
		out = append(out, documentToDTO(item))
	}
	return out, nil
}

func (s *DocumentService) GetStatus(ctx context.Context, id uint64) (*dto.DocumentResponse, error) {
	doc, err := s.docRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := documentToDTO(*doc)
	return &resp, nil
}

func (s *DocumentService) Delete(ctx context.Context, id uint64) error {
	doc, err := s.docRepo.Get(ctx, id)
	if err != nil {
		return err
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

func documentToDTO(doc model.Document) dto.DocumentResponse {
	return dto.DocumentResponse{
		ID:               doc.ID,
		KBID:             doc.KBID,
		Filename:         doc.Filename,
		OriginalFilename: doc.OriginalFilename,
		FileExt:          doc.FileExt,
		MimeType:         doc.MimeType,
		FileSize:         doc.FileSize,
		FileHash:         doc.FileHash,
		StoragePath:      doc.StoragePath,
		IndexStatus:      doc.IndexStatus,
		IndexError:       doc.IndexError,
		ChunkCount:       doc.ChunkCount,
		CreatedAt:        doc.CreatedAt,
		UpdatedAt:        doc.UpdatedAt,
	}
}

func safeStoredName(filename string) string {
	base := filepath.Base(filename)
	if base == "." || base == "" {
		return "upload"
	}
	return base
}
