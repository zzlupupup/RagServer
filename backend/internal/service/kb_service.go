package service

import (
	"context"
	"fmt"
	"strings"

	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/repository"
)

type KnowledgeBaseService struct {
	repo  *repository.KnowledgeBaseRepository
	users *repository.UserRepository
}

func NewKnowledgeBaseService(repo *repository.KnowledgeBaseRepository, users *repository.UserRepository) *KnowledgeBaseService {
	return &KnowledgeBaseService{repo: repo, users: users}
}

func (s *KnowledgeBaseService) Create(ctx context.Context, user *model.User, req dto.KnowledgeBaseCreateRequest) (*dto.KnowledgeBaseResponse, error) {
	visibility := model.VisibilityPrivate
	if user.Role == model.RoleTeacher {
		visibility = model.VisibilityPublic
	}
	kb := &model.KnowledgeBase{
		OwnerUserID: user.ID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Visibility:  visibility,
		Status:      model.StatusActive,
	}
	if kb.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if err := s.repo.Create(ctx, kb); err != nil {
		return nil, err
	}
	resp, err := s.toDTO(ctx, *kb, user.ID)
	return &resp, err
}

func (s *KnowledgeBaseService) List(ctx context.Context, user *model.User) ([]dto.KnowledgeBaseResponse, error) {
	items, err := s.repo.ListVisible(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.KnowledgeBaseResponse, 0, len(items))
	for _, item := range items {
		resp, err := s.toDTO(ctx, item, user.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, resp)
	}
	return out, nil
}

func (s *KnowledgeBaseService) Get(ctx context.Context, user *model.User, id uint64) (*dto.KnowledgeBaseResponse, error) {
	kb, err := s.repo.GetVisible(ctx, id, user.ID)
	if err != nil {
		return nil, err
	}
	resp, err := s.toDTO(ctx, *kb, user.ID)
	return &resp, err
}

func (s *KnowledgeBaseService) GetVisibleModel(ctx context.Context, user *model.User, id uint64) (*model.KnowledgeBase, error) {
	return s.repo.GetVisible(ctx, id, user.ID)
}

func (s *KnowledgeBaseService) ResolveVisibleByName(ctx context.Context, user *model.User, name string) (*model.KnowledgeBase, error) {
	items, err := s.repo.FindVisibleByName(ctx, user.ID, strings.TrimSpace(name))
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("knowledge base not found: %s", name)
	}
	if len(items) > 1 {
		return nil, fmt.Errorf("knowledge base name is ambiguous: %s", name)
	}
	return &items[0], nil
}

func (s *KnowledgeBaseService) Update(ctx context.Context, user *model.User, id uint64, req dto.KnowledgeBaseUpdateRequest) (*dto.KnowledgeBaseResponse, error) {
	kb, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if kb.OwnerUserID != user.ID {
		return nil, fmt.Errorf("only owner can update knowledge base")
	}
	if req.Name != nil {
		kb.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		kb.Description = strings.TrimSpace(*req.Description)
	}
	if err := s.repo.Update(ctx, kb); err != nil {
		return nil, err
	}
	resp, err := s.toDTO(ctx, *kb, user.ID)
	return &resp, err
}

func (s *KnowledgeBaseService) Delete(ctx context.Context, user *model.User, id uint64) error {
	kb, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if kb.OwnerUserID != user.ID {
		return fmt.Errorf("only owner can delete knowledge base")
	}
	return s.repo.Delete(ctx, id)
}

func (s *KnowledgeBaseService) CanView(ctx context.Context, user *model.User, kbID uint64) (*model.KnowledgeBase, error) {
	return s.repo.GetVisible(ctx, kbID, user.ID)
}

func (s *KnowledgeBaseService) toDTO(ctx context.Context, kb model.KnowledgeBase, userID uint64) (dto.KnowledgeBaseResponse, error) {
	count, err := s.repo.CountDocuments(ctx, kb.ID)
	if err != nil {
		return dto.KnowledgeBaseResponse{}, err
	}
	resp := dto.KnowledgeBaseResponse{
		ID:            kb.ID,
		OwnerUserID:   kb.OwnerUserID,
		Name:          kb.Name,
		Description:   kb.Description,
		Visibility:    kb.Visibility,
		Status:        kb.Status,
		DocumentCount: count,
		CanManage:     kb.OwnerUserID == userID,
		CreatedAt:     kb.CreatedAt,
		UpdatedAt:     kb.UpdatedAt,
	}
	if owner, _ := s.users.Get(ctx, kb.OwnerUserID); owner != nil {
		resp.OwnerUserDisplayName = owner.DisplayName
	}
	return resp, nil
}
