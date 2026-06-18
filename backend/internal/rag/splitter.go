package rag

import (
	"strings"

	"github.com/cloudwego/eino/schema"
)

type Splitter struct {
	ChunkSize int
	Overlap   int
}

func NewSplitter(chunkSize, overlap int) Splitter {
	if chunkSize <= 0 {
		chunkSize = 800
	}
	if overlap < 0 || overlap >= chunkSize {
		overlap = 120
	}
	return Splitter{ChunkSize: chunkSize, Overlap: overlap}
}

func (s Splitter) Split(docs []*schema.Document) []*schema.Document {
	var chunks []*schema.Document
	for _, doc := range docs {
		text := strings.TrimSpace(doc.Content)
		if text == "" {
			continue
		}
		runes := []rune(text)
		for start := 0; start < len(runes); {
			end := start + s.ChunkSize
			if end > len(runes) {
				end = len(runes)
			}
			meta := map[string]any{}
			for k, v := range doc.MetaData {
				meta[k] = v
			}
			chunks = append(chunks, &schema.Document{
				Content:  strings.TrimSpace(string(runes[start:end])),
				MetaData: meta,
			})
			if end == len(runes) {
				break
			}
			start = end - s.Overlap
		}
	}
	return chunks
}
