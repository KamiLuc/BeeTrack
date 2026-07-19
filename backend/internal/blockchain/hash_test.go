package blockchain

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestSHA256File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pdf")
	if err := os.WriteFile(path, []byte("hello honey"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	// sha256("hello honey")
	want := "80b1b71d335ba6e63e4d02a1f1d6801fe3bcbbc532b9c2833720855011d35105"

	got, err := SHA256File(path)
	if err != nil {
		t.Fatalf("SHA256File() error = %v", err)
	}
	if gotHex := hex.EncodeToString(got[:]); gotHex != want {
		t.Errorf("SHA256File() = %s, want %s", gotHex, want)
	}
}

func TestSHA256File_MissingFile(t *testing.T) {
	_, err := SHA256File(filepath.Join(t.TempDir(), "does-not-exist.pdf"))
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
