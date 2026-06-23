package service

import (
	"context"
	"fmt"
	"time"

	appcrypto "ragserver/backend/internal/crypto"
	"ragserver/backend/internal/dto"
	apperrors "ragserver/backend/internal/errors"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/repository"
)

type APIKeyService struct {
	repo   *repository.APIKeyRepository
	users  *repository.UserRepository
	secret string
}

type MCPIdentity struct {
	Key  *model.APIKey
	User *model.User
}

func NewAPIKeyService(repo *repository.APIKeyRepository, users *repository.UserRepository, secret string) *APIKeyService {
	return &APIKeyService{repo: repo, users: users, secret: secret}
}

func (s *APIKeyService) Create(ctx context.Context, teacher *model.User, req dto.APIKeyCreateRequest) (*dto.APIKeyResponse, error) {
	if teacher.Role != model.RoleTeacher {
		return nil, fmt.Errorf("teacher role required")
	}
	boundUser, err := s.users.Get(ctx, req.BoundUserID)
	if err != nil {
		return nil, fmt.Errorf("bound user not found")
	}
	if boundUser.Status != model.StatusActive {
		return nil, fmt.Errorf("bound user is disabled")
	}
	exists, err := s.repo.ExistsByCreatorAndBoundUser(ctx, teacher.ID, boundUser.ID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("student already has an api key")
	}
	plain, err := appcrypto.GenerateAPIKey()
	if err != nil {
		return nil, err
	}
	encrypted, err := appcrypto.EncryptString(s.secret, plain)
	if err != nil {
		return nil, err
	}
	key := &model.APIKey{
		CreatedByUserID: teacher.ID,
		BoundUserID:     boundUser.ID,
		Name:            req.Name,
		KeyHash:         appcrypto.HashAPIKey(s.secret, plain),
		EncryptedKey:    encrypted,
		Status:          model.StatusActive,
	}
	if err := s.repo.Create(ctx, key); err != nil {
		return nil, err
	}
	resp := apiKeyToDTO(*key, boundUser)
	resp.APIKey = plain
	return &resp, nil
}

func (s *APIKeyService) List(ctx context.Context, teacher *model.User, page, pageSize int) (*dto.PaginationResponse[dto.APIKeyResponse], error) {
	page, pageSize = normalizePagination(page, pageSize)
	keys, total, err := s.repo.ListByCreator(ctx, teacher.ID, page, pageSize)
	if err != nil {
		return nil, err
	}
	out := make([]dto.APIKeyResponse, 0, len(keys))
	for _, key := range keys {
		bound, _ := s.users.Get(ctx, key.BoundUserID)
		out = append(out, apiKeyToDTO(key, bound))
	}
	return &dto.PaginationResponse[dto.APIKeyResponse]{
		Items:    out,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *APIKeyService) Get(ctx context.Context, teacher *model.User, id uint64) (*dto.APIKeyResponse, error) {
	key, err := s.repo.GetByCreator(ctx, id, teacher.ID)
	if err != nil {
		return nil, err
	}
	bound, _ := s.users.Get(ctx, key.BoundUserID)
	resp := apiKeyToDTO(*key, bound)
	return &resp, nil
}

func (s *APIKeyService) Reveal(ctx context.Context, teacher *model.User, id uint64) (string, error) {
	key, err := s.repo.GetByCreator(ctx, id, teacher.ID)
	if err != nil {
		return "", err
	}
	return appcrypto.DecryptString(s.secret, key.EncryptedKey)
}

func (s *APIKeyService) Disable(ctx context.Context, teacher *model.User, id uint64) error {
	key, err := s.repo.GetByCreator(ctx, id, teacher.ID)
	if err != nil {
		return err
	}
	key.Status = model.StatusDisabled
	return s.repo.Update(ctx, key)
}

func (s *APIKeyService) Enable(ctx context.Context, teacher *model.User, id uint64) error {
	key, err := s.repo.GetByCreator(ctx, id, teacher.ID)
	if err != nil {
		return err
	}
	key.Status = model.StatusActive
	return s.repo.Update(ctx, key)
}

func (s *APIKeyService) Delete(ctx context.Context, teacher *model.User, id uint64) error {
	key, err := s.repo.GetByCreator(ctx, id, teacher.ID)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, key.ID)
}

func (s *APIKeyService) Authenticate(ctx context.Context, keyValue string) (*MCPIdentity, error) {
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
	user, err := s.users.Get(ctx, key.BoundUserID)
	if err != nil || user.Status != model.StatusActive {
		return nil, apperrors.ErrUnauthorized
	}
	now := time.Now()
	key.LastUsedAt = &now
	_ = s.repo.Update(ctx, key)
	return &MCPIdentity{Key: key, User: user}, nil
}

func apiKeyToDTO(key model.APIKey, bound *model.User) dto.APIKeyResponse {
	resp := dto.APIKeyResponse{
		ID:              key.ID,
		CreatedByUserID: key.CreatedByUserID,
		BoundUserID:     key.BoundUserID,
		Name:            key.Name,
		Status:          key.Status,
		LastUsedAt:      key.LastUsedAt,
		CreatedAt:       key.CreatedAt,
		UpdatedAt:       key.UpdatedAt,
	}
	if bound != nil {
		resp.BoundUserEmail = bound.Email
		resp.BoundUserDisplayName = bound.DisplayName
		resp.BoundUserRole = bound.Role
	}
	return resp
}
