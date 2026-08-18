package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAppLooksUpAnExactDictionaryEntry(t *testing.T) {
	app := &App{}
	path := filepath.Join("pkg", "dictionary", "testdata", "cc-cedict.mdx")
	if err := app.OpenDictionary(path); err != nil {
		t.Fatalf("open dictionary: %v", err)
	}
	defer app.CloseDictionary()

	entry, err := app.LookupWord("你好")
	if err != nil {
		t.Fatalf("lookup word: %v", err)
	}
	if !strings.Contains(entry.HTML, "hello") {
		t.Fatalf("unexpected entry HTML: %q", entry.HTML)
	}
}
