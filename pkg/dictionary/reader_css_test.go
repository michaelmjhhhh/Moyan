package dictionary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlineCSSResourcesEmbedsLocalFilesAndBlocksRemotePaths(t *testing.T) {
	directory := t.TempDir()
	stylesheet := filepath.Join(directory, "dictionary.css")
	if err := os.WriteFile(filepath.Join(directory, "font.woff2"), []byte("font-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(stylesheet, []byte(`@font-face { src: url("./font.woff2"); } .remote { background: url("https://example.com/x.png"); }`), 0o600); err != nil {
		t.Fatal(err)
	}

	css, err := inlineCSSResources(`@font-face { src: url("./font.woff2"); } .remote { background: url("https://example.com/x.png"); }`, stylesheet)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(css, "data:font/woff2;base64,") {
		t.Fatalf("expected local font to be inlined: %s", css)
	}
	if strings.Contains(css, "example.com") {
		t.Fatalf("expected remote URL to be removed: %s", css)
	}
}

func TestBoundedResourcePathRejectsTraversal(t *testing.T) {
	if _, err := boundedResourcePath("/tmp/dictionary", "../../secret.txt"); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}
