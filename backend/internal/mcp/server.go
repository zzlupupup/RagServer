package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/service"
)

type Server struct {
	apiKeys     *service.APIKeyService
	kbs         *service.KnowledgeBaseService
	documents   *service.DocumentService
	search      *service.SearchService
	uploadMaxMB int
	handler     http.Handler
}

func New(
	apiKeys *service.APIKeyService,
	kbs *service.KnowledgeBaseService,
	documents *service.DocumentService,
	search *service.SearchService,
	uploadMaxMB int,
) http.Handler {
	s := &Server{
		apiKeys:     apiKeys,
		kbs:         kbs,
		documents:   documents,
		search:      search,
		uploadMaxMB: uploadMaxMB,
	}
	mcpServer := sdk.NewServer(&sdk.Implementation{Name: "ragserver", Version: "0.1.0"}, nil)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        "kb.list",
		Description: "List all knowledge bases.",
	}, s.listKBs)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        "kb.upload_file",
		Description: "Upload a PDF, Markdown, or DOCX file to a knowledge base and index it.",
	}, s.uploadFile)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        "rag.search",
		Description: "Search a knowledge base using vector retrieval.",
	}, s.searchKB)

	streamable := sdk.NewStreamableHTTPHandler(func(req *http.Request) *sdk.Server {
		return mcpServer
	}, nil)
	s.handler = streamable
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r)
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if _, err := s.apiKeys.Authenticate(r.Context(), token); err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	s.handler.ServeHTTP(w, r)
}

func (s *Server) listKBs(ctx context.Context, req *sdk.CallToolRequest, args dto.MCPListKBRequest) (*sdk.CallToolResult, dto.KnowledgeBaseListResponse, error) {
	items, err := s.kbs.List(ctx)
	if err != nil {
		return nil, dto.KnowledgeBaseListResponse{}, err
	}
	payload := dto.KnowledgeBaseListResponse{Items: items}
	return textJSON(payload), payload, nil
}

func (s *Server) uploadFile(ctx context.Context, req *sdk.CallToolRequest, args dto.MCPUploadFileRequest) (*sdk.CallToolResult, dto.UploadFileResponse, error) {
	if args.KBID == 0 {
		return nil, dto.UploadFileResponse{}, fmt.Errorf("kb_id is required")
	}
	if strings.TrimSpace(args.Filename) == "" {
		return nil, dto.UploadFileResponse{}, fmt.Errorf("filename is required")
	}
	data, err := base64.StdEncoding.DecodeString(args.ContentBase64)
	if err != nil {
		return nil, dto.UploadFileResponse{}, fmt.Errorf("invalid content_base64: %w", err)
	}
	maxBytes := s.uploadMaxMB * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = 20 * 1024 * 1024
	}
	if len(data) > maxBytes {
		return nil, dto.UploadFileResponse{}, fmt.Errorf("file exceeds %dMB limit", maxBytes/1024/1024)
	}
	resp, err := s.documents.UploadBytes(ctx, args.KBID, args.Filename, args.MimeType, data)
	if err != nil {
		return nil, dto.UploadFileResponse{}, err
	}
	payload := dto.UploadFileResponse{
		KBID:        resp.KBID,
		DocumentID:  resp.ID,
		IndexStatus: resp.IndexStatus,
		ChunkCount:  resp.ChunkCount,
	}
	return textJSON(payload), payload, nil
}

func (s *Server) searchKB(ctx context.Context, req *sdk.CallToolRequest, args dto.MCPSearchRequest) (*sdk.CallToolResult, dto.SearchResponse, error) {
	if args.KBID == 0 {
		return nil, dto.SearchResponse{}, fmt.Errorf("kb_id is required")
	}
	resp, err := s.search.Search(ctx, args.KBID, args.Query, args.TopK)
	if err != nil {
		return nil, dto.SearchResponse{}, err
	}
	return textJSON(resp), *resp, nil
}

func textJSON(value any) *sdk.CallToolResult {
	data, _ := json.MarshalIndent(value, "", "  ")
	return &sdk.CallToolResult{
		Content: []sdk.Content{&sdk.TextContent{Text: string(data)}},
	}
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
