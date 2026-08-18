package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/michaelmjhhhh/Moyan/pkg/dictionary"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context

	mu      sync.RWMutex
	readers []*dictionary.Reader
	name    string
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(_ context.Context) {
	a.CloseDictionary()
}

func (a *App) OpenDictionary(path string) error {
	reader, err := dictionary.Open(path)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.readers = append(a.readers, reader)
	a.name = reader.Name()
	return nil
}

// ChooseAndOpenDictionary lets the native desktop shell choose one local MDX file.
// The file picker is intentionally not exposed as a browser-only capability.
func (a *App) ChooseAndOpenDictionary() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("application runtime is not ready")
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "导入词典",
		Filters: []runtime.FileFilter{
			{DisplayName: "MDX dictionary (*.mdx)", Pattern: "*.mdx"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := a.OpenDictionary(path); err != nil {
		return "", err
	}
	return a.name, nil
}

func (a *App) CurrentDictionary() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.name
}

func (a *App) CloseDictionary() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, reader := range a.readers {
		reader.Close()
	}
	a.readers = nil
	a.name = ""
}

func (a *App) LookupWord(word string) (dictionary.Entry, error) {
	entries, err := a.LookupWords(word)
	if err != nil {
		return dictionary.Entry{}, err
	}
	return entries[0], nil
}

func (a *App) SearchCandidates(word string) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	const perDictionaryLimit = 8
	var candidates []string
	seen := make(map[string]struct{})
	for _, reader := range a.readers {
		for _, candidate := range reader.Candidates(word, perDictionaryLimit) {
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

func (a *App) LookupWords(word string) ([]dictionary.Entry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	var entries []dictionary.Entry
	for _, reader := range a.readers {
		entry, err := reader.Lookup(word)
		if err == nil {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		return nil, dictionary.ErrNotFound
	}
	return entries, nil
}
