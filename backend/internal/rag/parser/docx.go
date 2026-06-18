package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func parseDOCX(data []byte) ([]*schema.Document, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var documentXML io.ReadCloser
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			documentXML, err = file.Open()
			if err != nil {
				return nil, err
			}
			break
		}
	}
	if documentXML == nil {
		return nil, fmt.Errorf("docx document.xml not found")
	}
	defer documentXML.Close()

	decoder := xml.NewDecoder(documentXML)
	var builder strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "p" && builder.Len() > 0 {
				builder.WriteByte('\n')
			}
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" {
				if builder.Len() > 0 && !strings.HasSuffix(builder.String(), "\n") {
					builder.WriteByte(' ')
				}
				builder.WriteString(text)
			}
		}
	}
	content := strings.TrimSpace(builder.String())
	if content == "" {
		return nil, fmt.Errorf("docx has no extractable text")
	}
	return []*schema.Document{{Content: content}}, nil
}
