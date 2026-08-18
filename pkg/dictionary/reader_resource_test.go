package dictionary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlineHTMLResourcesEmbedsLocalImageAndAudio(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "cover.png"), []byte("png-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "sound.mp3"), []byte("mp3-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	html := `<img src="cover.png"><audio controls src="sound.mp3"></audio><img src="https://example.com/remote.png">`
	got := inlineHTMLResources(html, directory, nil)
	if !strings.Contains(got, `src="data:image/png`) {
		t.Fatalf("expected local image data URL: %s", got)
	}
	if !strings.Contains(got, `src="data:audio/mpeg`) {
		t.Fatalf("expected local audio data URL: %s", got)
	}
	if strings.Contains(got, "example.com") {
		t.Fatalf("expected remote resource to be removed: %s", got)
	}
}
