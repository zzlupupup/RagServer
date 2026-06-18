package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	einoembedding "github.com/cloudwego/eino/components/embedding"
)

type OpenAIEmbedder struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAIEmbedder(baseURL, apiKey, model string) *OpenAIEmbedder {
	return &OpenAIEmbedder{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type embeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (e *OpenAIEmbedder) EmbedStrings(ctx context.Context, texts []string, opts ...einoembedding.Option) ([][]float64, error) {
	if len(texts) == 0 {
		return [][]float64{}, nil
	}
	body, err := json.Marshal(embeddingRequest{Model: e.model, Input: texts})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return nil, fmt.Errorf("embedding api error: %s", parsed.Error.Message)
		}
		return nil, fmt.Errorf("embedding api returned status %d", resp.StatusCode)
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("embedding api returned %d vectors, expected %d", len(parsed.Data), len(texts))
	}
	out := make([][]float64, len(texts))
	for i, item := range parsed.Data {
		idx := item.Index
		if idx < 0 || idx >= len(out) {
			idx = i
		}
		out[idx] = item.Embedding
	}
	return out, nil
}
