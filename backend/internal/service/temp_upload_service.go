package service

import (
	"context"
	"mime/multipart"

	"ragserver/backend/internal/dto"
	appstorage "ragserver/backend/internal/storage"
)

type TempUploadService struct {
	fileStore *appstorage.FileStore
	maxBytes  int64
}

func NewTempUploadService(fileStore *appstorage.FileStore, maxMB int) *TempUploadService {
	if maxMB <= 0 {
		maxMB = 20
	}
	return &TempUploadService{fileStore: fileStore, maxBytes: int64(maxMB) * 1024 * 1024}
}

func (s *TempUploadService) Upload(ctx context.Context, file *multipart.FileHeader) (*dto.TempUploadResponse, error) {
	path, filename, mimeType, size, err := s.fileStore.SaveTemp(file, s.maxBytes)
	if err != nil {
		return nil, err
	}
	return &dto.TempUploadResponse{
		FilePath: path,
		Filename: filename,
		MimeType: mimeType,
		FileSize: size,
	}, nil
}
