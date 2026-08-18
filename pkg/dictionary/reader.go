package dictionary

import (
	"encoding/base64"
	"errors"
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	mdictcore "github.com/michaelmjhhhh/Moyan/internal/mdictcore"
)

var (
	ErrUnsupportedFormat   = errors.New("unsupported dictionary format")
	ErrProtectedDictionary = errors.New("protected dictionaries are not supported")
	ErrNotFound            = errors.New("headword not found")

	maxStylesheetBytes       = 2 * 1024 * 1024
	maxResourceBytes   int64 = 8 * 1024 * 1024

	cssURLPattern = regexp.MustCompile(`(?i)url\(\s*["']?([^"')]+)["']?\s*\)`)
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
		inlined, err := inlineCSSResources(string(data), filepath.Join(directory, name))
		if err != nil {
			return "", err
		}
		styles.WriteString(inlined)
	}
	return styles.String(), nil
}

func inlineCSSResources(css, stylesheetPath string) (string, error) {
	directory := filepath.Dir(stylesheetPath)
	return cssURLPattern.ReplaceAllStringFunc(css, func(match string) string {
		submatches := cssURLPattern.FindStringSubmatch(match)
		if len(submatches) != 2 {
			return `url("")`
		}
		reference := strings.TrimSpace(submatches[1])
		if reference == "" || strings.HasPrefix(strings.ToLower(reference), "data:") {
			return match
		}
		lower := strings.ToLower(reference)
		if strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:") || strings.HasPrefix(reference, "//") {
			return `url("")`
		}
		parsed, err := url.Parse(reference)
		if err != nil || parsed.Path == "" {
			return `url("")`
		}
		resourcePath, err := boundedResourcePath(directory, parsed.Path)
		if err != nil {
			return `url("")`
		}
		info, err := os.Lstat(resourcePath)
		if err != nil || !info.Mode().IsRegular() || info.Size() > maxResourceBytes {
			return `url("")`
		}
		data, err := os.ReadFile(resourcePath)
		if err != nil || int64(len(data)) > maxResourceBytes {
			return `url("")`
		}
		contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(resourcePath)))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		return `url("data:` + contentType + `;base64,` + encoded + `")`
	}), nil
}

func boundedResourcePath(directory, reference string) (string, error) {
	decoded, err := url.PathUnescape(reference)
	if err != nil {
		return "", err
	}
	decoded = strings.ReplaceAll(decoded, "\\", "/")
	decoded = strings.TrimLeft(decoded, "/")
	candidate := filepath.Clean(filepath.Join(directory, filepath.FromSlash(decoded)))
	relative, err := filepath.Rel(directory, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return "", errors.New("resource escapes dictionary directory")
	}
	return candidate, nil
}

func (r *Reader) Close() {
	if r != nil {
		r.mdx = nil
	}
}
