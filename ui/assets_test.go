package ui

import (
	iofs "io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every generated chunk must survive embedding, including underscore-prefixed
// dynamic route chunks. A successful Vite/Go build alone did not catch this.
func TestGeneratedAssetsAreEmbedded(t *testing.T) {
	count := 0
	err := filepath.WalkDir("dist", func(path string, entry iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, ".html") {
			return nil
		}
		disk, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		embedded, err := GetUIAssets().ReadFile(filepath.ToSlash(path))
		if err != nil {
			t.Errorf("missing embedded asset %s: %v", path, err)
			return nil
		}
		if string(disk) != string(embedded) {
			t.Errorf("stale embedded asset %s", path)
		}
		count++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("build the UI before testing the embedded binary")
	}
}
