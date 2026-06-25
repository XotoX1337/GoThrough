package configstore

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/XotoX1337/GoThrough/config"
)

// TestEmbeddedConfigsAreValid loads every bundled YAML directly. List() skips
// invalid configs silently (so the picker stays usable), which means a broken
// bundled file would just vanish — this test fails loudly instead.
func TestEmbeddedConfigsAreValid(t *testing.T) {
	count := 0
	err := fs.WalkDir(embedded, "configs", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return err
		}
		data, rerr := embedded.ReadFile(path)
		if rerr != nil {
			t.Errorf("%s: read: %v", path, rerr)
			return nil
		}
		if _, lerr := config.LoadBytes(data); lerr != nil {
			t.Errorf("%s: %v", path, lerr)
		}
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if count == 0 {
		t.Fatal("no embedded configs found")
	}
}
