package embedding

import (
	"context"
	"fmt"
	"strings"
	"time"

	einoembedding "github.com/cloudwego/eino/components/embedding"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	arkmodel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type ArkMultimodalEmbedder struct {
	client *arkruntime.Client
	model  string
}

func NewArkMultimodalEmbedder(baseURL, apiKey, model string) *ArkMultimodalEmbedder {
	return &ArkMultimodalEmbedder{
		client: arkruntime.NewClientWithApiKey(
			apiKey,
			arkruntime.WithBaseUrl(normalizeArkBaseURL(baseURL)),
			arkruntime.WithTimeout(60*time.Second),
		),
		model: model,
	}
}

func (e *ArkMultimodalEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...einoembedding.Option) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}
	out := make([][]float64, len(texts))
	for i, text := range texts {
		vector, err := e.embedText(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("ark multimodal embedding text %d: %w", i, err)
		}
		out[i] = vector
	}
	return out, nil
}

func (e *ArkMultimodalEmbedder) embedText(ctx context.Context, text string) ([]float64, error) {
	resp, err := e.client.CreateMultiModalEmbeddings(ctx, arkmodel.MultiModalEmbeddingRequest{
		Model: e.model,
		Input: []arkmodel.MultimodalEmbeddingInput{
			{
				Type: arkmodel.MultiModalEmbeddingInputTypeText,
				Text: &text,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data.Embedding) == 0 {
		return nil, fmt.Errorf("embedding api returned no vector")
	}
	return float32sToFloat64s(resp.Data.Embedding), nil
}

func normalizeArkBaseURL(baseURL string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return "https://ark.cn-beijing.volces.com/api/v3"
	}
	if strings.HasSuffix(base, "/api/v3") {
		return base
	}
	return base + "/api/v3"
}

func float32sToFloat64s(values []float32) []float64 {
	out := make([]float64, len(values))
	for i, value := range values {
		out[i] = float64(value)
	}
	return out
}
