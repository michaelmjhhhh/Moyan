package dictionary_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/michaelmjhhhh/Moyan/pkg/dictionary"
)

func TestReaderLooksUpAnExactHeadword(t *testing.T) {
	path := filepath.Join("testdata", "cc-cedict.mdx")

	reader, err := dictionary.Open(path)
	if err != nil {
		t.Fatalf("open dictionary: %v", err)
	}
	defer reader.Close()

	entry, err := reader.Lookup("你好")
	if err != nil {
		t.Fatalf("lookup exact headword: %v", err)
	}
	if !strings.Contains(entry.HTML, "<p class=\"cc_def\">hello</p>") {
		t.Fatalf("lookup returned unexpected entry HTML: %q", entry.HTML)
	}
	if !strings.Contains(entry.CSS, ".cc_wrapper") {
		t.Fatalf("lookup did not load the sibling stylesheet: %q", entry.CSS)
	}
}
