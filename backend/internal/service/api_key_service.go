package service

import (
	"context"
	"time"

	appcrypto "ragserver/backend/internal/crypto"
	"ragserver/backend/internal/dto"
	apperrors "ragserver/backend/internal/errors"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/repository"
)

type APIKeyService struct {
	repo   *repository.APIKeyRepository
	secret string
}

func NewAPIKeyService(repo *repository.APIKeyRepository, secret string) *APIKeyService {
	return &APIKeyService{repo: repo, secret: secret}
}

func (s *APIKeyService) Create(ctx context.Context, name string) (*dto.APIKeyResponse, error) {
	plain, err := appcrypto.GenerateAPIKey()
	if err != nil {
		return nil, err
	}
	encrypted, err := appcrypto.EncryptString(s.secret, plain)
	if err != nil {
		return nil, err
	}
	key := &model.APIKey{
		Name:         name,
		KeyHash:      appcrypto.HashAPIKey(s.secret, plain),
		EncryptedKey: encrypted,
		Status:       model.StatusActive,
	}
	if err := s.repo.Create(ctx, key); err != nil {
		return nil, err
	}
	resp := apiKeyToDTO(*key)
	resp.APIKey = plain
	return &resp, nil
}

func (s *APIKeyService) List(ctx context.Context) ([]dto.APIKeyResponse, error) {
	keys, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.APIKeyResponse, 0, len(keys))
	for _, key := range keys {
		out = append(out, apiKeyToDTO(key))
	}
	return out, nil
}

func (s *APIKeyService) Get(ctx context.Context, id uint64) (*dto.APIKeyResponse, error) {
	key, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := apiKeyToDTO(*key)
	return &resp, nil
}

func (s *APIKeyService) Reveal(ctx context.Context, id uint64) (string, error) {
	key, err := s.repo.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return appcrypto.DecryptString(s.secret, key.EncryptedKey)
}

func (s *APIKeyService) Disable(ctx context.Context, id uint64) error {
	key, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	key.Status = model.StatusDisabled
	return s.repo.Update(ctx, key)
}

func (s *APIKeyService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *APIKeyService) Authenticate(ctx context.Context, keyValue string) (*model.APIKey, error) {
	if err := appcrypto.ValidateKeyPrefix(keyValue); err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	key, err := s.repo.GetByHash(ctx, appcrypto.HashAPIKey(s.secret, keyValue))
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	if key.Status != model.StatusActive {
		return nil, apperrors.ErrUnauthorized
	}
	now := time.Now()
	key.LastUsedAt = &now
	_ = s.repo.Update(ctx, key)
	return key, nil
}

func apiKeyToDTO(key model.APIKey) dto.APIKeyResponse {
	return dto.APIKeyResponse{
		ID:         key.ID,
		Name:       key.Name,
		Status:     key.Status,
		LastUsedAt: key.LastUsedAt,
		CreatedAt:  key.CreatedAt,
		UpdatedAt:  key.UpdatedAt,
	}
}
