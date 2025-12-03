package filelist

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/alx99/sail/internal/collator"
	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/style"
)

type stubSel struct{}

func (stubSel) IsSelected(string) bool { return false }

func TestChDirOrdersLikeLs(t *testing.T) {
	dir := t.TempDir()

	names := []string{
		"b",
		"A",
		".hidden",
		"a1",
		"a10",
		"a2",
		"Z",
		"zeta",
		"foo bar",
		".dotfile",
	}

	for _, name := range names {
		path := filepath.Join(dir, name)
		if strings.Contains(name, "bar") {
			if err := os.Mkdir(path, 0o755); err != nil {
				t.Fatalf("create dir %q: %v", name, err)
			}
			continue
		}
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			t.Fatalf("create file %q: %v", name, err)
		}
	}

	d, err := filesys.NewDir(dir)
	if err != nil {
		t.Fatalf("NewDir failed: %v", err)
	}

	coll := collator.New()

	expected := slices.Clone(d.Entries())
	sortEntries(coll, expected)

	v := New(dir, State{}, stubSel{}, coll, false, style.NewStyles(""))
	v.SetShowHidden(true)
	v.ChDir(d, State{})

	assertSameOrder(t, expected, v.entries)
}

func assertSameOrder(t *testing.T, expected, actual []filesys.DirEntry) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Fatalf("entry count mismatch: expected %d, got %d", len(expected), len(actual))
	}

	for i := range expected {
		if expected[i].Name() != actual[i].Name() || expected[i].IsDir() != actual[i].IsDir() {
			t.Fatalf("ordering mismatch at %d: expected %q (dir=%v), got %q (dir=%v)",
				i, expected[i].Name(), expected[i].IsDir(), actual[i].Name(), actual[i].IsDir())
		}
	}
}
