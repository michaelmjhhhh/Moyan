package dictionary

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mdictcore "github.com/michaelmjhhhh/Moyan/internal/mdictcore"
)

var (
	ErrUnsupportedFormat   = errors.New("unsupported dictionary format")
	ErrProtectedDictionary = errors.New("protected dictionaries are not supported")
	ErrNotFound            = errors.New("headword not found")

	maxStylesheetBytes = 2 * 1024 * 1024
)

type Entry struct {
	Headword string
	HTML     string
	CSS      string
}

type Reader struct {
	mdx        *mdictcore.Mdict
	stylesheet string
}

func Open(path string) (*Reader, error) {
	mdx, err := mdictcore.New(path)
	if err != nil {
		return nil, fmt.Errorf("open MDX: %w", err)
	}
	if mdx.IsMDD() || strings.EqualFold(filepath.Ext(path), ".mdd") {
		return nil, fmt.Errorf("%w: MDD is a resource package, not a dictionary", ErrUnsupportedFormat)
	}
	if mdx.IsRecordEncrypted() {
		return nil, ErrProtectedDictionary
	}
	if err := mdx.BuildIndex(); err != nil {
		return nil, fmt.Errorf("build MDX index: %w", err)
	}
	stylesheet, err := loadStylesheets(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("load dictionary CSS: %w", err)
	}
	return &Reader{mdx: mdx, stylesheet: stylesheet}, nil
}

func (r *Reader) Name() string {
	if r == nil || r.mdx == nil {
		return ""
	}
	return r.mdx.Name()
}

func (r *Reader) Lookup(headword string) (Entry, error) {
	if r == nil || r.mdx == nil {
		return Entry{}, errors.New("dictionary reader is closed")
	}
	trimmed := strings.TrimSpace(headword)
	if trimmed == "" {
		return Entry{}, ErrNotFound
	}
	html, err := r.mdx.Lookup(trimmed)
	if err != nil {
		return Entry{}, fmt.Errorf("%w: %s", ErrNotFound, trimmed)
	}
	return Entry{
		Headword: trimmed,
		HTML:     strings.TrimRight(string(html), "\x00"),
		CSS:      r.stylesheet,
	}, nil
}

func loadStylesheets(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".css") {
			continue
		}
		info, err := os.Lstat(filepath.Join(directory, entry.Name()))
		if err != nil {
			return "", err
		}
		if !info.Mode().IsRegular() {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	var styles strings.Builder
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			return "", err
		}
		if len(data) > maxStylesheetBytes || styles.Len()+len(data) > maxStylesheetBytes {
			return "", fmt.Errorf("stylesheet exceeds %d bytes", maxStylesheetBytes)
		}
		if styles.Len() > 0 {
			styles.WriteString("\n")
		}
		styles.Write(data)
	}
	return styles.String(), nil
}

func (r *Reader) Close() {
	if r != nil {
		r.mdx = nil
	}
}
