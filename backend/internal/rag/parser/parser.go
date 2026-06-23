package parser

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(ctx context.Context, filename string, data []byte) ([]*schema.Document, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	var (
		docs []*schema.Document
		err  error
	)
	switch ext {
	case ".md", ".markdown":
		docs = parseMarkdown(data)
	case ".pdf":
		docs, err = parsePDF(ctx, bytes.NewReader(data))
	case ".docx":
		docs, err = parseDOCX(data)
	case ".doc":
		return nil, fmt.Errorf(".doc files are not supported, please convert to .docx")
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
	if err != nil {
		return nil, err
	}
	return normalizeDocuments(ext, docs)
}

func SupportedExtension(filename string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".pdf", ".md", ".markdown", ".docx":
		return true
	default:
		return false
	}
}

func normalizeDocuments(ext string, docs []*schema.Document) ([]*schema.Document, error) {
	out := make([]*schema.Document, 0, len(docs))
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		content := strings.TrimSpace(doc.Content)
		if content == "" {
			continue
		}
		normalized := *doc
		normalized.Content = content
		out = append(out, &normalized)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s has no extractable text", strings.TrimPrefix(ext, "."))
	}
	return out, nil
}
