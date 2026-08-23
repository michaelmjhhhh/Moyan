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

	mu     sync.RWMutex
	reader *dictionary.Reader
	name   string
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
	if a.reader != nil {
		a.reader.Close()
	}
	a.reader = reader
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
		Title: "Import dictionary",
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
	return a.CurrentDictionary(), nil
}

func (a *App) CurrentDictionary() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.name
}

func (a *App) CloseDictionary() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.reader != nil {
		a.reader.Close()
		a.reader = nil
	}
	a.name = ""
}

func (a *App) LookupWord(word string) (dictionary.Entry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.reader == nil {
		return dictionary.Entry{}, dictionary.ErrNotFound
	}
	return a.reader.Lookup(word)
}

func (a *App) SearchCandidates(word string) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.reader == nil {
		return nil
	}
	return a.reader.Candidates(word, 8)
}
