package dictionary

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	mdictcore "github.com/michaelmjhhhh/Moyan/internal/mdictcore"
)

var (
	ErrUnsupportedFormat   = errors.New("unsupported dictionary format")
	ErrProtectedDictionary = errors.New("protected dictionaries are not supported")
	ErrNotFound            = errors.New("headword not found")
)

type Entry struct {
	Headword string
	HTML     string
}

type Reader struct {
	mdx *mdictcore.Mdict
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
	return &Reader{mdx: mdx}, nil
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
	return Entry{Headword: trimmed, HTML: strings.TrimRight(string(html), "\x00")}, nil
}

func (r *Reader) Close() {
	if r != nil {
		r.mdx = nil
	}
}
