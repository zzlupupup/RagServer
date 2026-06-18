package parser

import (
	"context"
	"io"

	pdfparser "github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/schema"
)

func parsePDF(ctx context.Context, reader io.Reader) ([]*schema.Document, error) {
	p, err := pdfparser.NewPDFParser(ctx, &pdfparser.Config{ToPages: false})
	if err != nil {
		return nil, err
	}
	return p.Parse(ctx, reader)
}
