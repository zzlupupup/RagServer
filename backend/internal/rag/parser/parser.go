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
	switch ext {
	case ".md", ".markdown":
		return parseMarkdown(data), nil
	case ".pdf":
		return parsePDF(ctx, bytes.NewReader(data))
	case ".docx":
		return parseDOCX(data)
	case ".doc":
		return nil, fmt.Errorf(".doc files are not supported, please convert to .docx")
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func SupportedExtension(filename string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".pdf", ".md", ".markdown", ".docx":
		return true
	default:
		return false
	}
}
