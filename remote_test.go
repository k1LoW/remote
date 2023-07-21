package remote

import (
	"bytes"
	"io"
	"testing"
)

func TestOpenHTTP(t *testing.T) {
	f, err := Open("https://example.com")
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() {
		defer f.Close()
	})
	b, err := io.ReadAll(f)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(b, []byte("Example Domain")) {
		t.Error("invalid content")
	}
}

func TestOpenGitHub(t *testing.T) {
	f, err := Open("github://k1LoW/gh-setup/README.md")
	if err != nil {
		t.Error(err)
	}
	t.Cleanup(func() {
		defer f.Close()
	})
	b, err := io.ReadAll(f)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(b, []byte("gh-setup")) {
		t.Error("invalid content")
	}
}
