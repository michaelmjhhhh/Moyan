package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelmjhhhh/Moyan/pkg/dictionary"
)

func openFixture(t *testing.T) *App {
	t.Helper()
	app := &App{}
	path := filepath.Join("pkg", "dictionary", "testdata", "cc-cedict.mdx")
	if err := app.OpenDictionary(path); err != nil {
		t.Fatalf("open dictionary: %v", err)
	}
	t.Cleanup(app.CloseDictionary)
	return app
}

func TestAppLookupWithoutDictionary(t *testing.T) {
	app := &App{}
	_, err := app.LookupWord("你好")
	if !errors.Is(err, dictionary.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestAppLooksUpAnExactDictionaryEntry(t *testing.T) {
	app := openFixture(t)

	entry, err := app.LookupWord("你好")
	if err != nil {
		t.Fatalf("lookup word: %v", err)
	}
	if !strings.Contains(entry.HTML, "hello") {
		t.Fatalf("unexpected entry HTML: %q", entry.HTML)
	}
}

func TestAppReplacingADictionaryKeepsExactLookup(t *testing.T) {
	app := openFixture(t)
	path := filepath.Join("pkg", "dictionary", "testdata", "cc-cedict.mdx")
	if err := app.OpenDictionary(path); err != nil {
		t.Fatalf("replace dictionary: %v", err)
	}

	if got := app.CurrentDictionary(); got == "" {
		t.Fatal("expected current dictionary name after replace")
	}
	entry, err := app.LookupWord("你好")
	if err != nil {
		t.Fatalf("lookup word: %v", err)
	}
	if !strings.Contains(entry.HTML, "hello") {
		t.Fatalf("unexpected entry HTML: %q", entry.HTML)
	}
}

func TestAppFailedOpenKeepsCurrentDictionary(t *testing.T) {
	app := openFixture(t)
	if err := app.OpenDictionary(filepath.Join("pkg", "dictionary", "testdata", "missing.mdx")); err == nil {
		t.Fatal("expected open of missing file to fail")
	}

	entry, err := app.LookupWord("你好")
	if err != nil {
		t.Fatalf("lookup after failed open: %v", err)
	}
	if !strings.Contains(entry.HTML, "hello") {
		t.Fatalf("unexpected entry HTML: %q", entry.HTML)
	}
}

func TestAppLookupMissingHeadword(t *testing.T) {
	app := openFixture(t)
	_, err := app.LookupWord("this-headword-is-not-in-the-fixture")
	if !errors.Is(err, dictionary.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestAppSearchCandidatesAfterOpen(t *testing.T) {
	app := openFixture(t)
	candidates := app.SearchCandidates("你")
	for _, candidate := range candidates {
		if candidate == "你好" {
			return
		}
	}
	t.Fatalf("expected 你好 candidate, got %#v", candidates)
}
