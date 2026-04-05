package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Document represents a documentation file.
type Document struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Path  string `json:"-"`
}

// SystemService handles system-level operations like reading documentation.
type SystemService struct {
	docsDir string
}

// NewSystemService creates a new SystemService.
func NewSystemService(docsDir string) *SystemService {
	return &SystemService{docsDir: docsDir}
}

// ListDocs returns a list of all documentation files in the docs directory.
func (s *SystemService) ListDocs() ([]Document, error) {
	var docs []Document

	err := filepath.Walk(s.docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			relPath, err := filepath.Rel(s.docsDir, path)
			if err != nil {
				return err
			}

			// Create a slug by replacing slashes and extensions
			slug := strings.TrimSuffix(relPath, ".md")
			slug = strings.ReplaceAll(slug, "/", "-")

			// Simple title: capitalize and replace hyphens/underscores
			title := strings.TrimSuffix(info.Name(), ".md")
			title = strings.ReplaceAll(title, "_", " ")
			title = strings.ReplaceAll(title, "-", " ")
			title = strings.Title(title)

			docs = append(docs, Document{
				Slug:  slug,
				Title: title,
				Path:  relPath,
			})
		}
		return nil
	})

	return docs, err
}

// GetDocContent returns the raw content of a documentation file by its slug.
func (s *SystemService) GetDocContent(slug string) (string, error) {
	// 1. Convert slug back to relative path (simple mapping for now)
	relPath := strings.ReplaceAll(slug, "-", "/") + ".md"

	// 2. Security Check: Block directory traversal
	cleanedPath := filepath.Clean(relPath)
	if strings.HasPrefix(cleanedPath, "..") || filepath.IsAbs(cleanedPath) {
		return "", fmt.Errorf("invalid path: sequence detected")
	}

	// 3. Construct absolute path
	absPath := filepath.Join(s.docsDir, cleanedPath)

	// 4. Verify path is still within docsDir
	if !strings.HasPrefix(absPath, filepath.Clean(s.docsDir)) {
		return "", fmt.Errorf("access denied")
	}

	// 5. Read file
	content, err := os.ReadFile(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("document not found: %s", slug)
		}
		return "", err
	}

	return string(content), nil
}
