package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelmjhhhh/Moyan/pkg/dictionary"
)

func fixturePath() string {
	return filepath.Join("pkg", "dictionary", "testdata", "cc-cedict.mdx")
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	app := &App{libraryFile: filepath.Join(t.TempDir(), "library.json")}
	t.Cleanup(app.CloseDictionary)
	return app
}

func openFixture(t *testing.T) *App {
	t.Helper()
	app := newTestApp(t)
	if err := app.OpenDictionary(fixturePath()); err != nil {
		t.Fatalf("open dictionary: %v", err)
	}
	return app
}

func copyFixture(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(fixturePath())
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "cc-cedict-copy.mdx")
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatalf("write fixture copy: %v", err)
	}
	return dst
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
	if got := len(app.Library()); got != 1 {
		t.Fatalf("library size after failed open: %d", got)
	}

	entry, err := app.LookupWord("你好")
	if err != nil {
		t.Fatalf("lookup after failed open: %v", err)
	}
	if !strings.Contains(entry.HTML, "hello") {
		t.Fatalf("unexpected entry HTML: %q", entry.HTML)
	}
}

func TestAppKeepsTwoOpenPackages(t *testing.T) {
	app := newTestApp(t)
	first := fixturePath()
	second := copyFixture(t)
	if err := app.OpenDictionary(first); err != nil {
		t.Fatalf("open first: %v", err)
	}
	if err := app.OpenDictionary(second); err != nil {
		t.Fatalf("open second: %v", err)
	}
	pkgs := app.Library()
	if len(pkgs) != 2 {
		t.Fatalf("library size: %d", len(pkgs))
	}
	if pkgs[0].Path == pkgs[1].Path {
		t.Fatal("expected distinct package paths")
	}
}

func TestAppLookupInQueriesEachPackageIndependently(t *testing.T) {
	app := newTestApp(t)
	first := fixturePath()
	second := copyFixture(t)
	if err := app.OpenDictionary(first); err != nil {
		t.Fatalf("open first: %v", err)
	}
	if err := app.OpenDictionary(second); err != nil {
		t.Fatalf("open second: %v", err)
	}

	e1, err := app.LookupIn(first, "你好")
	if err != nil {
		t.Fatalf("lookup first: %v", err)
	}
	e2, err := app.LookupIn(second, "你好")
	if err != nil {
		t.Fatalf("lookup second: %v", err)
	}
	if !strings.Contains(e1.HTML, "hello") {
		t.Fatalf("first entry HTML: %q", e1.HTML)
	}
	if !strings.Contains(e2.HTML, "hello") {
		t.Fatalf("second entry HTML: %q", e2.HTML)
	}
}

func TestAppRestoresLibraryOnNewProcess(t *testing.T) {
	store := filepath.Join(t.TempDir(), "library.json")
	first := &App{libraryFile: store}
	if err := first.OpenDictionary(fixturePath()); err != nil {
		t.Fatalf("open: %v", err)
	}
	first.CloseDictionary()

	second := &App{libraryFile: store}
	t.Cleanup(second.CloseDictionary)
	if err := second.RestoreLibrary(); err != nil {
		t.Fatalf("restore: %v", err)
	}
	pkgs := second.Library()
	if len(pkgs) == 0 {
		t.Fatal("restored 词库 is empty")
	}
	entry, err := second.LookupIn(pkgs[0].Path, "你好")
	if err != nil {
		t.Fatalf("lookup after restore: %v", err)
	}
	if !strings.Contains(entry.HTML, "hello") {
		t.Fatalf("restored entry HTML: %q", entry.HTML)
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
