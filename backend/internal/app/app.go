package app

import (
	"context"
	"net/http"
	"time"

	"ragserver/backend/internal/api"
	"ragserver/backend/internal/config"
	mcpserver "ragserver/backend/internal/mcp"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/rag"
	ragembedding "ragserver/backend/internal/rag/embedding"
	"ragserver/backend/internal/repository"
	"ragserver/backend/internal/service"
	"ragserver/backend/internal/storage"

	"github.com/gin-gonic/gin"
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
		&model.User{},
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
	userRepo := repository.NewUserRepository(db)
	docRepo := repository.NewDocumentRepository(db)
	chunkRepo := repository.NewChunkRepository(db)
	jobRepo := repository.NewIngestionJobRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	recoverStaleIndexing(ctx, docRepo, jobRepo, cfg.IndexTimeoutSeconds)

	embedder, err := ragembedding.NewEmbedder(cfg)
	if err != nil {
		return nil, err
	}
	pipeline := rag.NewPipeline(redisClient, embedder)
	fileStore := storage.NewFileStore(cfg.FileStorageDir, cfg.MCPTmpDir)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiresHours)
	userSvc := service.NewUserService(userRepo)
	kbSvc := service.NewKnowledgeBaseService(kbRepo, userRepo)
	apiKeySvc := service.NewAPIKeyService(apiKeyRepo, userRepo, cfg.APIKeyEncryptionSecret)
	docSvc := service.NewDocumentService(kbRepo, docRepo, chunkRepo, jobRepo, fileStore, pipeline, userRepo, cfg.IndexTimeoutSeconds)
	searchSvc := service.NewSearchService(kbRepo, pipeline)
	tempUploadSvc := service.NewTempUploadService(fileStore, cfg.MCPUploadMaxMB)

	handler := &api.Handler{
		Auth:        authSvc,
		Users:       userSvc,
		KB:          kbSvc,
		Document:    docSvc,
		APIKey:      apiKeySvc,
		Search:      searchSvc,
		TempUploads: tempUploadSvc,
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	registerRoutes(r, handler, authSvc)
	mcpHandler := mcpserver.New(apiKeySvc, kbSvc, docSvc, searchSvc, cfg.MCPUploadMaxMB)
	r.Any("/mcp", gin.WrapH(mcpHandler))
	return &App{Config: cfg, Router: r}, nil
}

func (a *App) Run() error {
	return http.ListenAndServe(a.Config.ServerAddr, a.Router)
}

func recoverStaleIndexing(ctx context.Context, docs *repository.DocumentRepository, jobs *repository.IngestionJobRepository, timeoutSeconds int) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}
	olderThan := time.Now().Add(-time.Duration(timeoutSeconds) * time.Second)
	message := "indexing interrupted or timed out"
	_ = docs.FailStaleIndexing(ctx, olderThan, message)
	_ = jobs.FailStaleRunning(ctx, olderThan, message)
}
