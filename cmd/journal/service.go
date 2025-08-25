package journal

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sumwatshade/surflog/cmd/create"
)

// Service defines persistence operations for journal entries.
type Service interface {
	List() ([]create.Entry, error)
	Get(id string) (create.Entry, error)
	Create(e create.Entry) (create.Entry, error)
	Update(id string, mutate func(*create.Entry) error) (create.Entry, error)
}

var _ Service = (*fileService)(nil)

// fileService stores each entry as a JSON file under baseDir.
type fileService struct {
	baseDir string
}

// NewFileService creates a journal service rooted at dir (created if missing).
func NewFileService(dir string) (Service, error) {
	if dir == "" {
		return nil, errors.New("empty journal dir")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &fileService{baseDir: dir}, nil
}

func (s *fileService) entryPath(id string) string { return filepath.Join(s.baseDir, id+".json") }

// List loads all entry JSON files (best-effort; skips corrupt ones) sorted by mtime desc.
func (s *fileService) List() ([]create.Entry, error) {
	var entries []create.Entry
	// gather files
	var files []fs.FileInfo
	dir, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}
	for _, de := range dir {
		if de.IsDir() || !strings.HasSuffix(de.Name(), ".json") {
			continue
		}
		info, err := de.Info()
		if err != nil { // skip
			continue
		}
		files = append(files, info)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].ModTime().After(files[j].ModTime()) })
	for _, fi := range files {
		b, err := os.ReadFile(filepath.Join(s.baseDir, fi.Name()))
		if err != nil {
			continue
		}
		var e create.Entry
		if err := json.Unmarshal(b, &e); err != nil || e.ID == "" {
			continue
		}
		if strings.TrimSpace(e.CreatedAt) == "" { // backfill from file mtime
			e.CreatedAt = fi.ModTime().UTC().Format(time.RFC3339)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *fileService) Get(id string) (create.Entry, error) {
	if id == "" {
		return create.Entry{}, errors.New("empty id")
	}
	b, err := os.ReadFile(s.entryPath(id))
	if err != nil {
		return create.Entry{}, err
	}
	var e create.Entry
	if err := json.Unmarshal(b, &e); err != nil {
		return create.Entry{}, err
	}
	if e.ID == "" {
		return create.Entry{}, errors.New("entry missing id")
	}
	return e, nil
}

func (s *fileService) Create(e create.Entry) (create.Entry, error) {
	e.ID = uuid.NewString()
	if strings.TrimSpace(e.Spot) == "" {
		return create.Entry{}, errors.New("spot required")
	}
	if strings.TrimSpace(e.CreatedAt) == "" {
		e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return create.Entry{}, err
	}
	if err := os.WriteFile(s.entryPath(e.ID), data, 0o644); err != nil {
		return create.Entry{}, err
	}
	return e, nil
}

func (s *fileService) Update(id string, mutate func(*create.Entry) error) (create.Entry, error) {
	cur, err := s.Get(id)
	if err != nil {
		return create.Entry{}, err
	}
	if mutate != nil {
		if err := mutate(&cur); err != nil {
			return create.Entry{}, err
		}
	}
	cur.ID = id                                 // safety
	if strings.TrimSpace(cur.CreatedAt) == "" { // ensure not lost
		cur.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	data, err := json.MarshalIndent(cur, "", "  ")
	if err != nil {
		return create.Entry{}, err
	}
	tmp := s.entryPath(id) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return create.Entry{}, err
	}
	if err := os.Rename(tmp, s.entryPath(id)); err != nil {
		return create.Entry{}, err
	}
	return cur, nil
}
