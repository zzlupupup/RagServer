package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileStore struct {
	Root string
}

func NewFileStore(root string) *FileStore {
	return &FileStore{Root: root}
}

func (s *FileStore) Save(kbID, documentID uint64, originalName string, data []byte) (string, string, error) {
	cleanName := sanitizeFilename(originalName)
	rel := filepath.Join(fmt.Sprintf("%d", kbID), time.Now().Format("2006-01-02"), fmt.Sprintf("%d", documentID), cleanName)
	full := filepath.Join(s.Root, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		return "", "", err
	}
	if err := os.WriteFile(full, data, 0644); err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(data)
	return full, hex.EncodeToString(sum[:]), nil
}

func (s *FileStore) Delete(path string) error {
	if path == "" {
		return nil
	}
	return os.Remove(path)
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	if name == "." || name == "" {
		return "upload.bin"
	}
	return name
}
