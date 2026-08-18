package main

import (
	"context"
	"sync"

	"github.com/michaelmjhhhh/Moyan/pkg/dictionary"
)

type App struct {
	ctx context.Context

	mu     sync.RWMutex
	reader *dictionary.Reader
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
	return nil
}

func (a *App) CloseDictionary() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.reader != nil {
		a.reader.Close()
		a.reader = nil
	}
}

func (a *App) LookupWord(word string) (dictionary.Entry, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.reader == nil {
		return dictionary.Entry{}, dictionary.ErrNotFound
	}
	return a.reader.Lookup(word)
}
