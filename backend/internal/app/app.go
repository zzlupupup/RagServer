package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"ragserver/backend/internal/api"
	"ragserver/backend/internal/config"
	mcpserver "ragserver/backend/internal/mcp"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/rag"
	ragembedding "ragserver/backend/internal/rag/embedding"
	"ragserver/backend/internal/repository"
	"ragserver/backend/internal/service"
	"ragserver/backend/internal/storage"
)

type App struct {
	Config config.Config
	Router *gin.Engine
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := storage.OpenMySQL(cfg.MySQLDSN)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&model.KnowledgeBase{},
		&model.Document{},
		&model.DocumentChunk{},
		&model.IngestionJob{},
		&model.APIKey{},
	); err != nil {
		return nil, err
	}

	redisClient := storage.OpenRedis(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	if err := storage.EnsureVectorIndex(ctx, redisClient, cfg.EmbeddingDimension); err != nil {
		return nil, err
	}

	kbRepo := repository.NewKnowledgeBaseRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	chunkRepo := repository.NewChunkRepository(db)
	jobRepo := repository.NewIngestionJobRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)

	embedder := ragembedding.NewOpenAIEmbedder(cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.EmbeddingModel)
	pipeline := rag.NewPipeline(redisClient, embedder)
	fileStore := storage.NewFileStore(cfg.FileStorageDir)

	kbSvc := service.NewKnowledgeBaseService(kbRepo)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, cfg.APIKeyEncryptionSecret)
	docSvc := service.NewDocumentService(kbRepo, docRepo, chunkRepo, jobRepo, fileStore, pipeline)
	searchSvc := service.NewSearchService(kbRepo, pipeline)

	handler := &api.Handler{
		KB:       kbSvc,
		Document: docSvc,
		APIKey:   apiKeySvc,
		Search:   searchSvc,
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	registerRoutes(r, handler, cfg.AdminToken)
	mcpHandler := mcpserver.New(apiKeySvc, kbSvc, docSvc, searchSvc, cfg.MCPUploadMaxMB)
	r.Any("/mcp", gin.WrapH(mcpHandler))

	return &App{Config: cfg, Router: r}, nil
}

func (a *App) Run() error {
	return http.ListenAndServe(a.Config.ServerAddr, a.Router)
}
