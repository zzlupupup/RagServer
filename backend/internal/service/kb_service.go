package service

import (
	"context"
	"strings"

	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/repository"
)

type KnowledgeBaseService struct {
	repo *repository.KnowledgeBaseRepository
}

func NewKnowledgeBaseService(repo *repository.KnowledgeBaseRepository) *KnowledgeBaseService {
	return &KnowledgeBaseService{repo: repo}
}

func (s *KnowledgeBaseService) Create(ctx context.Context, req dto.KnowledgeBaseCreateRequest) (*dto.KnowledgeBaseResponse, error) {
	kb := &model.KnowledgeBase{
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Status:      model.StatusActive,
	}
	if err := s.repo.Create(ctx, kb); err != nil {
		return nil, err
	}
	resp, err := s.toDTO(ctx, *kb)
	return &resp, err
}

func (s *KnowledgeBaseService) List(ctx context.Context) ([]dto.KnowledgeBaseResponse, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.KnowledgeBaseResponse, 0, len(items))
	for _, item := range items {
		resp, err := s.toDTO(ctx, item)
		if err != nil {
			return nil, err
		}
		out = append(out, resp)
	}
	return out, nil
}

func (s *KnowledgeBaseService) Get(ctx context.Context, id uint64) (*dto.KnowledgeBaseResponse, error) {
	kb, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	resp, err := s.toDTO(ctx, *kb)
	return &resp, err
}

func (s *KnowledgeBaseService) Update(ctx context.Context, id uint64, req dto.KnowledgeBaseUpdateRequest) (*dto.KnowledgeBaseResponse, error) {
	kb, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
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
	resp, err := s.toDTO(ctx, *kb)
	return &resp, err
}

func (s *KnowledgeBaseService) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *KnowledgeBaseService) toDTO(ctx context.Context, kb model.KnowledgeBase) (dto.KnowledgeBaseResponse, error) {
	count, err := s.repo.CountDocuments(ctx, kb.ID)
	if err != nil {
		return dto.KnowledgeBaseResponse{}, err
	}
	return dto.KnowledgeBaseResponse{
		ID:            kb.ID,
		Name:          kb.Name,
		Description:   kb.Description,
		Status:        kb.Status,
		DocumentCount: count,
		CreatedAt:     kb.CreatedAt,
		UpdatedAt:     kb.UpdatedAt,
	}, nil
}
