package parser

import (
	"regexp"
	"strings"

	"github.com/cloudwego/eino/schema"
)

func parseMarkdown(data []byte) []*schema.Document {
	text := string(data)
	replacements := []struct {
		pattern string
		replace string
	}{
		{`(?m)^#{1,6}\s*`, ""},
		{`\*\*([^*]+)\*\*`, "$1"},
		{`\*([^*]+)\*`, "$1"},
		{"`([^`]+)`", "$1"},
		{`!\[([^\]]*)\]\([^)]+\)`, "$1"},
		{`\[([^\]]+)\]\([^)]+\)`, "$1"},
	}
	for _, item := range replacements {
		text = regexp.MustCompile(item.pattern).ReplaceAllString(text, item.replace)
	}
	return []*schema.Document{{Content: strings.TrimSpace(text)}}
}
