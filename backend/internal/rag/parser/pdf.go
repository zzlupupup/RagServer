package parser

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	pdfparser "github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/schema"
)

// einoParseTimeout bounds the eino-ext (dslipak/pdf) parser. That library's
// lexer can fall into a CPU-bound infinite loop on certain malformed PDFs and
// ignores context cancellation, so a context deadline alone cannot interrupt
// it — we must run it in a goroutine and abandon it on timeout.
const einoParseTimeout = 60 * time.Second

func parsePDF(ctx context.Context, reader io.Reader) ([]*schema.Document, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	// Prefer pdftotext (poppler): fast, robust, and context-bounded via
	// exec.CommandContext. Fall back to the eino parser only if pdftotext is
	// unavailable or yields no text.
	pdftotextDocs, pdftotextErr := parsePDFWithPdftotext(ctx, data)
	if pdftotextErr == nil && hasExtractableText(pdftotextDocs) {
		return pdftotextDocs, nil
	}
	einoDocs, einoErr := parsePDFWithEino(ctx, data)
	if einoErr == nil && hasExtractableText(einoDocs) {
		return einoDocs, nil
	}
	if pdftotextErr == nil && pdftotextDocs != nil {
		return pdftotextDocs, nil
	}
	if einoErr == nil && einoDocs != nil {
		return einoDocs, nil
	}
	return nil, fmt.Errorf("pdf parsing failed (pdftotext: %v; eino: %v)", pdftotextErr, einoErr)
}

func parsePDFWithEino(ctx context.Context, data []byte) ([]*schema.Document, error) {
	type result struct {
		docs []*schema.Document
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		p, err := pdfparser.NewPDFParser(ctx, &pdfparser.Config{ToPages: false})
		if err != nil {
			ch <- result{nil, err}
			return
		}
		docs, err := p.Parse(ctx, bytes.NewReader(data))
		ch <- result{docs, err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(einoParseTimeout):
		// Abandon the goroutine; it cannot be interrupted, but this request no
		// longer blocks on it.
		return nil, fmt.Errorf("eino pdf parser timed out after %s", einoParseTimeout)
	case r := <-ch:
		return r.docs, r.err
	}
}

func parsePDFWithPdftotext(ctx context.Context, data []byte) ([]*schema.Document, error) {
	file, err := os.CreateTemp("", "ragserver-pdf-*.pdf")
	if err != nil {
		return nil, err
	}
	path := file.Name()
	defer os.Remove(path)
	if _, err := file.Write(data); err != nil {
		file.Close()
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	output, err := exec.CommandContext(ctx, "pdftotext", "-layout", "-enc", "UTF-8", path, "-").Output()
	if err != nil {
		return nil, fmt.Errorf("pdftotext failed: %w", err)
	}
	text := strings.TrimSpace(string(output))
	if text == "" {
		return nil, nil
	}
	return []*schema.Document{{Content: text}}, nil
}

func hasExtractableText(docs []*schema.Document) bool {
	for _, doc := range docs {
		if doc != nil && strings.TrimSpace(doc.Content) != "" {
			return true
		}
	}
	return false
}
