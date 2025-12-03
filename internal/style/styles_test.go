package style

import (
	"io/fs"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MockDirEntry implements fs.DirEntry for testing
type MockDirEntry struct {
	name  string
	isDir bool
	mode  fs.FileMode
}

func (m MockDirEntry) Name() string               { return m.name }
func (m MockDirEntry) IsDir() bool                { return m.isDir }
func (m MockDirEntry) Type() fs.FileMode          { return m.mode.Type() }
func (m MockDirEntry) Info() (fs.FileInfo, error) { return MockFileInfo{m}, nil }

type MockFileInfo struct{ e MockDirEntry }

func (m MockFileInfo) Name() string       { return m.e.name }
func (m MockFileInfo) Size() int64        { return 0 }
func (m MockFileInfo) Mode() fs.FileMode  { return m.e.mode }
func (m MockFileInfo) ModTime() time.Time { return time.Now() }
func (m MockFileInfo) IsDir() bool        { return m.e.isDir }
func (m MockFileInfo) Sys() any           { return nil }

func TestGetStyle_Precedence(t *testing.T) {
	// Simulate LS_COLORS where 'no' (normal) is defined before 'di' (directory)
	// This causes the bug where directories get 'no' style if 'no' matches everything.
	lsColors := "no=00:di=01;34:*.txt=01;33"
	styles := NewStyles(lsColors)

	// Test Directory "1brc"
	dir := MockDirEntry{name: "1brc", isDir: true, mode: fs.ModeDir | 0o755}
	style := styles.GetStyle(dir)

	// Expected: Blue (4)
	// 'no=00' returns empty style (no foreground set).

	expectedColor := lipgloss.Color("4")
	actualColor := style.GetForeground()

	if actualColor != expectedColor {
		t.Errorf("Directory should have color Blue (4), but got %v. Likely matched 'no' instead of 'di'.", actualColor)
	}
}
