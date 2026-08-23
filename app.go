package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/michaelmjhhhh/Moyan/pkg/dictionary"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type PackageInfo struct {
	Path string
	Name string
}

type packageSlot struct {
	path   string
	reader *dictionary.Reader
}

type libraryFile struct {
	Paths []string `json:"paths"`
}

type App struct {
	ctx context.Context

	libraryFile string

	mu       sync.RWMutex
	packages []packageSlot
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	_ = a.RestoreLibrary()
}

func (a *App) shutdown(_ context.Context) {
	a.CloseDictionary()
}

func (a *App) libraryPath() string {
	if a.libraryFile != "" {
		return a.libraryFile
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "Moyan", "library.json")
}

func (a *App) persistLocked() error {
	path := a.libraryPath()
	if path == "" {
		return nil
	}
	paths := make([]string, 0, len(a.packages))
	for _, slot := range a.packages {
		paths = append(paths, slot.path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(libraryFile{Paths: paths})
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (a *App) RestoreLibrary() error {
	path := a.libraryPath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var file libraryFile
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	for _, item := range file.Paths {
		_ = a.openDictionary(item, false)
	}
	return nil
}

func (a *App) OpenDictionary(path string) error {
	return a.openDictionary(path, true)
}

func (a *App) openDictionary(path string, persist bool) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	reader, err := dictionary.Open(abs)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	for _, slot := range a.packages {
		if slot.path == abs {
			reader.Close()
			return nil
		}
	}
	a.packages = append(a.packages, packageSlot{path: abs, reader: reader})
	if persist {
		return a.persistLocked()
	}
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

func (a *App) Library() []PackageInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]PackageInfo, 0, len(a.packages))
	for _, slot := range a.packages {
		name := ""
		if slot.reader != nil {
			name = slot.reader.Name()
		}
		out = append(out, PackageInfo{Path: slot.path, Name: name})
	}
	return out
}

func (a *App) CurrentDictionary() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.packages) == 0 || a.packages[len(a.packages)-1].reader == nil {
		return ""
	}
	return a.packages[len(a.packages)-1].reader.Name()
}

func (a *App) CloseDictionary() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, slot := range a.packages {
		if slot.reader != nil {
			slot.reader.Close()
		}
	}
	a.packages = nil
}

func (a *App) LookupIn(path, word string) (dictionary.Entry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	abs, err := filepath.Abs(path)
	if err != nil {
		return dictionary.Entry{}, err
	}
	for _, slot := range a.packages {
		if slot.path == abs && slot.reader != nil {
			return slot.reader.Lookup(word)
		}
	}
	return dictionary.Entry{}, dictionary.ErrNotFound
}

func (a *App) LookupWord(word string) (dictionary.Entry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if len(a.packages) == 0 || a.packages[0].reader == nil {
		return dictionary.Entry{}, dictionary.ErrNotFound
	}
	return a.packages[0].reader.Lookup(word)
}

func (a *App) SearchCandidates(word string) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	const limit = 8
	var candidates []string
	seen := make(map[string]struct{})
	for _, slot := range a.packages {
		if slot.reader == nil {
			continue
		}
		for _, candidate := range slot.reader.Candidates(word, limit) {
			if _, ok := seen[candidate]; ok {
				continue
			}
			seen[candidate] = struct{}{}
			candidates = append(candidates, candidate)
			if len(candidates) >= limit {
				return candidates
			}
		}
	}
	return candidates
}
