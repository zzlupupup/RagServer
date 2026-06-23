package embedding

import (
	"fmt"
	"strings"

	einoembedding "github.com/cloudwego/eino/components/embedding"
	"ragserver/backend/internal/config"
)

func NewEmbedder(cfg config.Config) (einoembedding.Embedder, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.EmbeddingProvider)) {
	case "", "openai", "openai_compatible":
		return NewOpenAIEmbedder(cfg.OpenAIBaseURL, cfg.OpenAIAPIKey, cfg.EmbeddingModel), nil
	case "ark", "ark_multimodal", "volcengine_ark":
		apiKey := cfg.ArkAPIKey
		if apiKey == "" {
			apiKey = cfg.OpenAIAPIKey
		}
		return NewArkMultimodalEmbedder(cfg.ArkBaseURL, apiKey, cfg.EmbeddingModel), nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", cfg.EmbeddingProvider)
	}
}
