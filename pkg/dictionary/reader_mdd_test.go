package dictionary

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMDDResourceCanResolveCSSReference(t *testing.T) {
	mdds, err := loadMDDs(filepath.Join("testdata"))
	if err != nil {
		t.Fatal(err)
	}
	data, ok := lookupMDDResource(mdds, "\\._ccced.css")
	if !ok || len(data) == 0 {
		t.Fatalf("expected an MDD resource, got %d bytes", len(data))
	}

	css, err := inlineCSSResourcesWithLookup(`@import url("\\._ccced.css");`, filepath.Join("testdata", "dictionary.css"), func(reference string) ([]byte, bool) {
		return lookupMDDResource(mdds, reference)
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(css, "data:text/css") || !strings.Contains(css, ";base64,") {
		t.Fatalf("expected MDD CSS to be inlined: %s", css)
	}
}
