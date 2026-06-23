package service

import (
	"context"
	"fmt"
	"strings"

	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/rag"
	"ragserver/backend/internal/repository"
)

type SearchService struct {
	kbRepo   *repository.KnowledgeBaseRepository
	pipeline *rag.Pipeline
}

func NewSearchService(kbRepo *repository.KnowledgeBaseRepository, pipeline *rag.Pipeline) *SearchService {
	return &SearchService{kbRepo: kbRepo, pipeline: pipeline}
}

func (s *SearchService) Search(ctx context.Context, user *model.User, kbID uint64, query string, topK int) (*dto.SearchResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	kb, err := s.kbRepo.GetVisible(ctx, kbID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not visible")
	}
	items, err := s.pipeline.Search(ctx, kbID, query, topK)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].KBName = kb.Name
	}
	return &dto.SearchResponse{
		Summary: fmt.Sprintf("Found %d relevant chunks in %s.", len(items), kb.Name),
		Items:   items,
	}, nil
}
