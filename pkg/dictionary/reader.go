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

	cssURLPattern    = regexp.MustCompile(`(?i)url\(\s*["']?([^"')]+)["']?\s*\)`)
	styleLinkPattern = regexp.MustCompile(`(?i)<link\b[^>]*href=["']([^"']+)["'][^>]*>`)
)

type Entry struct {
	Dictionary string
	Headword   string
	HTML       string
	CSS        string
}

type Reader struct {
	mdx        *mdictcore.Mdict
	mdds       []*mdictcore.Mdict
	directory  string
	stylesheet string
	headwords  []string
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
	directory := filepath.Dir(path)
	mdds, err := loadMDDs(directory)
	if err != nil {
		return nil, fmt.Errorf("load dictionary resources: %w", err)
	}
	stylesheet, err := loadStylesheetsWithResolver(directory, func(reference string) ([]byte, bool) {
		return lookupMDDResource(mdds, reference)
	})
	if err != nil {
		return nil, fmt.Errorf("load dictionary CSS: %w", err)
	}
	keyEntries, err := mdx.GetKeyWordEntries()
	if err != nil {
		return nil, fmt.Errorf("read MDX headwords: %w", err)
	}
	headwords := make([]string, 0, len(keyEntries))
	for _, entry := range keyEntries {
		headwords = append(headwords, entry.KeyWord)
	}
	return &Reader{mdx: mdx, mdds: mdds, directory: directory, stylesheet: stylesheet, headwords: headwords}, nil
}

func (r *Reader) Candidates(query string, limit int) []string {
	if r == nil || r.mdx == nil || limit <= 0 {
		return nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil
	}
	queryLower := strings.ToLower(query)
	result := make([]string, 0, limit)
	seen := make(map[string]struct{}, limit)
	appendCandidate := func(word string) bool {
		if _, ok := seen[word]; ok {
			return len(result) >= limit
		}
		seen[word] = struct{}{}
		result = append(result, word)
		return len(result) >= limit
	}
	for _, word := range r.headwords {
		if strings.EqualFold(word, query) && appendCandidate(word) {
			return result
		}
	}
	for _, word := range r.headwords {
		if strings.HasPrefix(strings.ToLower(word), queryLower) && appendCandidate(word) {
			return result
		}
	}
	if len(result) < limit && len([]rune(query)) <= 32 {
		for _, word := range r.headwords {
			if distanceAtMost(strings.ToLower(word), queryLower, 2) && appendCandidate(word) {
				return result
			}
		}
	}
	return result
}

func distanceAtMost(left, right string, max int) bool {
	leftRunes, rightRunes := []rune(left), []rune(right)
	if absInt(len(leftRunes)-len(rightRunes)) > max {
		return false
	}
	previous := make([]int, len(rightRunes)+1)
	for i := range previous {
		previous[i] = i
	}
	for i, leftRune := range leftRunes {
		current := make([]int, len(rightRunes)+1)
		current[0] = i + 1
		rowMin := current[0]
		for j, rightRune := range rightRunes {
			cost := 0
			if leftRune != rightRune {
				cost = 1
			}
			current[j+1] = minInt(current[j]+1, previous[j+1]+1, previous[j]+cost)
			rowMin = minInt(rowMin, current[j+1])
		}
		if rowMin > max {
			return false
		}
		previous = current
	}
	return previous[len(rightRunes)] <= max
}

func minInt(values ...int) int {
	result := values[0]
	for _, value := range values[1:] {
		if value < result {
			result = value
		}
	}
	return result
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
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
	htmlText := strings.TrimRight(string(html), "\x00")
	htmlText = inlineHTMLResources(htmlText, r.directory, func(reference string) ([]byte, bool) {
		return lookupMDDResource(r.mdds, reference)
	})
	stylesheet := r.stylesheet + loadLinkedMDDStylesheets(htmlText, r.directory, r.mdds)
	return Entry{
		Dictionary: r.Name(),
		Headword:   trimmed,
		HTML:       htmlText,
		CSS:        stylesheet,
	}, nil
}

func loadStylesheets(directory string) (string, error) {
	return loadStylesheetsWithResolver(directory, nil)
}

func loadStylesheetsWithResolver(directory string, lookup resourceLookup) (string, error) {
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
		inlined, err := inlineCSSResourcesWithLookup(string(data), filepath.Join(directory, name), lookup)
		if err != nil {
			return "", err
		}
		styles.WriteString(inlined)
	}
	return styles.String(), nil
}

type resourceLookup func(reference string) ([]byte, bool)

func inlineCSSResources(css, stylesheetPath string) (string, error) {
	return inlineCSSResourcesWithLookup(css, stylesheetPath, nil)
}

func inlineCSSResourcesWithLookup(css, stylesheetPath string, lookup resourceLookup) (string, error) {
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
		dataURL, ok := resolveResourceDataURL(reference, directory, lookup)
		if !ok {
			return `url("")`
		}
		return `url("` + dataURL + `")`
	}), nil
}

var htmlResourcePattern = regexp.MustCompile(`(?i)(\b(?:src|poster)\s*=\s*["'])([^"']+)(["'])`)

func inlineHTMLResources(html, directory string, lookup resourceLookup) string {
	return htmlResourcePattern.ReplaceAllStringFunc(html, func(match string) string {
		submatches := htmlResourcePattern.FindStringSubmatch(match)
		if len(submatches) != 4 {
			return match
		}
		dataURL, ok := resolveResourceDataURL(submatches[2], directory, lookup)
		if !ok {
			return submatches[1] + "" + submatches[3]
		}
		return submatches[1] + dataURL + submatches[3]
	})
}

func resolveResourceDataURL(reference, directory string, lookup resourceLookup) (string, bool) {
	reference = strings.TrimSpace(reference)
	if reference == "" || strings.HasPrefix(strings.ToLower(reference), "data:") {
		return reference, reference != ""
	}
	lower := strings.ToLower(reference)
	if strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:") || strings.HasPrefix(reference, "//") {
		return "", false
	}
	parsed, err := url.Parse(reference)
	if err != nil || parsed.Path == "" {
		return "", false
	}
	var data []byte
	contentType := ""
	resourcePath, pathErr := boundedResourcePath(directory, parsed.Path)
	if pathErr == nil {
		info, statErr := os.Lstat(resourcePath)
		if statErr == nil && info.Mode().IsRegular() && info.Size() <= maxResourceBytes {
			data, _ = os.ReadFile(resourcePath)
			contentType = mime.TypeByExtension(strings.ToLower(filepath.Ext(resourcePath)))
		}
	}
	if len(data) == 0 && lookup != nil {
		data, _ = lookup(reference)
		contentType = mime.TypeByExtension(strings.ToLower(filepath.Ext(parsed.Path)))
	}
	if len(data) == 0 || int64(len(data)) > maxResourceBytes {
		return "", false
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data), true
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

func loadLinkedMDDStylesheets(html, directory string, mdds []*mdictcore.Mdict) string {
	seen := make(map[string]struct{})
	var styles strings.Builder
	for _, match := range styleLinkPattern.FindAllStringSubmatch(html, -1) {
		if len(match) != 2 {
			continue
		}
		reference := strings.TrimSpace(match[1])
		if _, ok := seen[reference]; ok {
			continue
		}
		seen[reference] = struct{}{}
		data, ok := lookupMDDResource(mdds, reference)
		if !ok || len(data) == 0 || len(data) > maxStylesheetBytes {
			continue
		}
		inlined, err := inlineCSSResourcesWithLookup(string(data), filepath.Join(directory, filepath.Base(reference)), func(resource string) ([]byte, bool) {
			return lookupMDDResource(mdds, resource)
		})
		if err != nil {
			continue
		}
		if styles.Len() > 0 {
			styles.WriteString("\n")
		}
		styles.WriteString(inlined)
	}
	if styles.Len() == 0 {
		return ""
	}
	return "\n" + styles.String()
}

func loadMDDs(directory string) ([]*mdictcore.Mdict, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".mdd") {
			continue
		}
		info, err := os.Lstat(filepath.Join(directory, entry.Name()))
		if err != nil {
			return nil, err
		}
		if info.Mode().IsRegular() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	mdds := make([]*mdictcore.Mdict, 0, len(names))
	for _, name := range names {
		mdd, err := mdictcore.New(filepath.Join(directory, name))
		if err != nil {
			return nil, fmt.Errorf("open %s: %w", name, err)
		}
		if err := mdd.BuildIndex(); err != nil {
			return nil, fmt.Errorf("index %s: %w", name, err)
		}
		mdds = append(mdds, mdd)
	}
	return mdds, nil
}

func lookupMDDResource(mdds []*mdictcore.Mdict, reference string) ([]byte, bool) {
	parsed, err := url.Parse(reference)
	if err != nil || parsed.Path == "" {
		return nil, false
	}
	path := strings.ReplaceAll(parsed.Path, "\\", "/")
	path = strings.TrimLeft(path, "/")
	candidates := []string{parsed.Path, "\\" + path, path}
	for _, mdd := range mdds {
		for _, candidate := range candidates {
			data, err := mdd.Lookup(candidate)
			if err == nil && len(data) > 0 {
				return data, true
			}
		}
	}
	return nil, false
}

func (r *Reader) Close() {
	if r != nil {
		r.mdx = nil
		r.mdds = nil
		r.headwords = nil
	}
}
