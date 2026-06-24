package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/service"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type identityKey struct{}

type Server struct {
	apiKeys   *service.APIKeyService
	kbs       *service.KnowledgeBaseService
	documents *service.DocumentService
	search    *service.SearchService
	handler   http.Handler
}

func New(
	apiKeys *service.APIKeyService,
	kbs *service.KnowledgeBaseService,
	documents *service.DocumentService,
	search *service.SearchService,
	uploadMaxMB int,
) http.Handler {
	s := &Server{apiKeys: apiKeys, kbs: kbs, documents: documents, search: search}
	mcpServer := sdk.NewServer(&sdk.Implementation{Name: "ragserver", Version: "0.1.0"}, nil)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        "kb.list",
		Description: "List knowledge bases visible to the API key bound user.",
	}, s.listKBs)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name: "kb.upload_file",
		Description: `Import a previously uploaded temporary file into a knowledge base. 
		- First run: curl -F "file=@localfilefullpath" http://120.55.76.32:80/api/v1/mcp/files/upload . 
		- Then pass the returned file_path and the target kb_name to this tool.`,
	}, s.uploadFile)
	sdk.AddTool(mcpServer, &sdk.Tool{
		Name:        "rag.search",
		Description: "Search a knowledge base by name using vector retrieval.",
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
	identity, err := s.apiKeys.Authenticate(r.Context(), token)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ctx := context.WithValue(r.Context(), identityKey{}, identity.User)
	s.handler.ServeHTTP(w, r.WithContext(ctx))
}

func (s *Server) listKBs(ctx context.Context, req *sdk.CallToolRequest, args dto.MCPListKBRequest) (*sdk.CallToolResult, dto.KnowledgeBaseListResponse, error) {
	user, err := boundUser(ctx)
	if err != nil {
		return nil, dto.KnowledgeBaseListResponse{}, err
	}
	items, err := s.kbs.List(ctx, user)
	if err != nil {
		return nil, dto.KnowledgeBaseListResponse{}, err
	}
	payload := dto.KnowledgeBaseListResponse{Items: items}
	return textJSON(payload), payload, nil
}

func (s *Server) uploadFile(ctx context.Context, req *sdk.CallToolRequest, args dto.MCPUploadFileRequest) (*sdk.CallToolResult, dto.UploadFileResponse, error) {
	user, err := boundUser(ctx)
	if err != nil {
		return nil, dto.UploadFileResponse{}, err
	}
	if strings.TrimSpace(args.KBName) == "" {
		return nil, dto.UploadFileResponse{}, fmt.Errorf("kb_name is required")
	}
	if strings.TrimSpace(args.FilePath) == "" {
		return nil, dto.UploadFileResponse{}, fmt.Errorf("file_path is required")
	}
	kb, err := s.kbs.ResolveVisibleByName(ctx, user, args.KBName)
	if err != nil {
		return nil, dto.UploadFileResponse{}, err
	}
	resp, err := s.documents.UploadTempFile(ctx, user, kb.ID, args.FilePath)
	if err != nil {
		return nil, dto.UploadFileResponse{}, err
	}
	payload := dto.UploadFileResponse{
		KBName:      kb.Name,
		KBID:        kb.ID,
		DocumentID:  resp.ID,
		IndexStatus: resp.IndexStatus,
		ChunkCount:  resp.ChunkCount,
	}
	return textJSON(payload), payload, nil
}

func (s *Server) searchKB(ctx context.Context, req *sdk.CallToolRequest, args dto.MCPSearchRequest) (*sdk.CallToolResult, dto.SearchResponse, error) {
	user, err := boundUser(ctx)
	if err != nil {
		return nil, dto.SearchResponse{}, err
	}
	if strings.TrimSpace(args.KBName) == "" {
		return nil, dto.SearchResponse{}, fmt.Errorf("kb_name is required")
	}
	kb, err := s.kbs.ResolveVisibleByName(ctx, user, args.KBName)
	if err != nil {
		return nil, dto.SearchResponse{}, err
	}
	resp, err := s.search.Search(ctx, user, kb.ID, args.Query, args.TopK)
	if err != nil {
		return nil, dto.SearchResponse{}, err
	}
	return textJSON(resp), *resp, nil
}

func textJSON(value any) *sdk.CallToolResult {
	data, _ := json.MarshalIndent(value, "", "  ")
	return &sdk.CallToolResult{Content: []sdk.Content{&sdk.TextContent{Text: string(data)}}}
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func boundUser(ctx context.Context) (*model.User, error) {
	user, ok := ctx.Value(identityKey{}).(*model.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("missing bound user")
	}
	return user, nil
}
