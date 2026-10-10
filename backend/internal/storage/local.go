// Package storage implements domain.FileStorage on the local filesystem.
package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/posumkm/backend/pkg/hash"
)

type Local struct {
	dir     string
	baseURL string
}

func NewLocal(dir, baseURL string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &Local{dir: dir, baseURL: strings.TrimRight(baseURL, "/")}, nil
}

func (l *Local) Save(_ context.Context, folder, ext string, r io.Reader) (string, error) {
	name, err := hash.RandomToken(16)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(l.dir, folder), 0o755); err != nil {
		return "", err
	}
	rel := filepath.ToSlash(filepath.Join(folder, name+ext))
	f, err := os.Create(filepath.Join(l.dir, rel))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return l.baseURL + "/" + rel, nil
}

// Delete removes a file previously returned by Save. URLs from elsewhere are ignored.
func (l *Local) Delete(_ context.Context, url string) error {
	rel, ok := strings.CutPrefix(url, l.baseURL+"/")
	if !ok || strings.Contains(rel, "..") {
		return nil
	}
	err := os.Remove(filepath.Join(l.dir, filepath.FromSlash(rel)))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
