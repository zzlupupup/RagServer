package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileStore struct {
	Root    string
	TmpRoot string
}

func NewFileStore(root, tmpRoot string) *FileStore {
	return &FileStore{Root: root, TmpRoot: tmpRoot}
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

func (s *FileStore) SaveTemp(file *multipart.FileHeader, maxBytes int64) (string, string, string, int64, error) {
	if maxBytes <= 0 {
		maxBytes = 20 * 1024 * 1024
	}
	if file.Size > maxBytes {
		return "", "", "", 0, fmt.Errorf("file exceeds %dMB limit", maxBytes/1024/1024)
	}
	name := sanitizeFilename(file.Filename)
	ext := strings.ToLower(filepath.Ext(name))
	if !supportedExt(ext) {
		return "", "", "", 0, fmt.Errorf("unsupported file type: %s", ext)
	}
	src, err := file.Open()
	if err != nil {
		return "", "", "", 0, err
	}
	defer src.Close()

	dir := filepath.Join(s.TmpRoot, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", "", 0, err
	}
	stored := "upload_" + uuid.NewString() + ext
	full := filepath.Join(dir, stored)
	dst, err := os.Create(full)
	if err != nil {
		return "", "", "", 0, err
	}
	defer dst.Close()
	written, err := io.Copy(dst, io.LimitReader(src, maxBytes+1))
	if err != nil {
		return "", "", "", 0, err
	}
	if written > maxBytes {
		_ = os.Remove(full)
		return "", "", "", 0, fmt.Errorf("file exceeds %dMB limit", maxBytes/1024/1024)
	}
	return full, name, file.Header.Get("Content-Type"), written, nil
}

func (s *FileStore) ReadTemp(path string) (string, []byte, error) {
	full, err := s.cleanTempPath(path)
	if err != nil {
		return "", nil, err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		return "", nil, err
	}
	return filepath.Base(full), data, nil
}

func (s *FileStore) DeleteTemp(path string) error {
	full, err := s.cleanTempPath(path)
	if err != nil {
		return err
	}
	return os.Remove(full)
}

func (s *FileStore) cleanTempPath(path string) (string, error) {
	full, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	root, err := filepath.Abs(s.TmpRoot)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, full)
	if err != nil || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("file_path must be under temporary upload directory")
	}
	return full, nil
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

func supportedExt(ext string) bool {
	switch ext {
	case ".pdf", ".md", ".markdown", ".docx":
		return true
	default:
		return false
	}
}
