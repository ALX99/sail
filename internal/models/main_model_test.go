package models

import (
	"errors"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/alx99/sail/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// create a struct that implements fs.DirEntry
type dirEntry struct {
	name      string
	isDir     bool
	isSymlink bool
	isRegular bool
	mode      fs.FileMode
	info      fs.FileInfo
}

var _ fs.DirEntry = dirEntry{}

func (d dirEntry) Name() string               { return d.name }
func (d dirEntry) IsDir() bool                { return d.isDir }
func (d dirEntry) Type() fs.FileMode          { return d.mode }
func (d dirEntry) Info() (fs.FileInfo, error) { return d.info, nil }

func TestModel_Update(t *testing.T) {
	pathAnimDuration = time.Duration(0)
	type fields struct {
		cfg                 config.Config
		cwd                 string
		files               []fs.DirEntry
		cursor              position
		cachedDirSelections map[string]string
		maxRows             int
		sb                  strings.Builder
		lastError           error
	}
	type args struct {
		msg tea.Msg
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		filesMock    []dirEntry
		filesErrMock error
		want         Model
		want1Nil     bool
		want1Result  tea.Msg
	}{
		{
			name: "Test special case",
			fields: fields{
				cwd:                 "/go",
				files:               []fs.DirEntry{},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: dirLoaded{
					path: "/go/src",
					files: []fs.DirEntry{
						dirEntry{name: "goa", isDir: true},
						dirEntry{name: "go", isDir: true},
					},
				},
			},
			want: Model{
				cwd: "/go/src",
				files: []fs.DirEntry{
					dirEntry{name: "goa", isDir: true},
					dirEntry{name: "go", isDir: true},
				},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
				prevCWD:             "/go",
			},
			want1Result: clearPrevCWD{},
		},
		{
			name: "Test cached filename exists",
			fields: fields{
				cwd:                 "/currpath",
				files:               []fs.DirEntry{},
				cursor:              position{},
				cachedDirSelections: map[string]string{"/cache": "ex"},
				maxRows:             10,
			},
			args: args{
				msg: dirLoaded{
					path: "/cache",
					files: []fs.DirEntry{
						dirEntry{name: "file1", isDir: true},
						dirEntry{name: "ex", isDir: true},
					},
				},
			},
			want: Model{
				cwd: "/cache",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: true},
					dirEntry{name: "ex", isDir: true},
				},
				cursor:              position{r: 1, c: 0},
				cachedDirSelections: map[string]string{"/cache": "ex"},
				maxRows:             10,
				prevCWD:             "/currpath",
			},
			want1Result: clearPrevCWD{},
		},
		{
			name: "Test adding directories to cachedDirSelections",
			fields: fields{
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "dir1", isDir: true},
				},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: dirLoaded{
					path:  "/testpath/dir1",
					files: []fs.DirEntry{},
				},
			},
			want: Model{
				cwd:                 "/testpath/dir1",
				files:               []fs.DirEntry{},
				cursor:              position{},
				cachedDirSelections: map[string]string{"/testpath": "dir1"},
				maxRows:             10,
				prevCWD:             "/testpath",
			},
			want1Result: clearPrevCWD{},
		},
		{
			name: "Test NavUp functionality",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavUp: "a"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 1, c: 0},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavUp: "a"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			want1Nil: true,
		},
		{
			name: "Test NavUp when cursor is at the top",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavUp: "a"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
					dirEntry{name: "file3", isDir: false},
				},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavUp: "a"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
					dirEntry{name: "file3", isDir: false},
				},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			want1Nil: true,
		},
		{
			name: "Test NavDown functionality",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavDown: "s"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavDown: "s"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 1, c: 0},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			want1Nil: true,
		},
		{
			name: "Test NavDown functionality when already at the bottom",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavDown: "s"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavDown: "s"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			want1Nil: true,
		},
		{
			name: "Test NavDown functionality when cursor is at the bottom of the current row",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavDown: "s"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
					dirEntry{name: "file3", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             2,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavDown: "s"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
					dirEntry{name: "file3", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             2,
			},
			want1Nil: true,
		},
		{
			name: "Test NavLeft",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavLeft: "a"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavLeft: "a"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 0},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			want1Nil: true,
		},
		{
			name: "Test NavRight",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavRight: "d"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 0},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavRight: "d"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			want1Nil: true,
		},
		{
			name: "Test NavRight when cursor is at the last column",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavRight: "d"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavRight: "d"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			want1Nil: true,
		},
		{
			name: "Test lastError is cleared on update",
			fields: fields{
				cfg:                 config.Config{},
				cwd:                 "/testpath",
				files:               []fs.DirEntry{},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
				lastError:           errors.New("previous error"),
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			},
			want: Model{
				cfg:                 config.Config{},
				cwd:                 "/testpath",
				files:               []fs.DirEntry{},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			want1Nil: true,
		},
		{
			name: "Load same directory",
			fields: fields{
				cwd:                 "/special",
				files:               []fs.DirEntry{},
				cursor:              position{},
				cachedDirSelections: map[string]string{"/special": "xx"},
				maxRows:             10,
			},
			args: args{
				msg: dirLoaded{
					path: "/special",
					files: []fs.DirEntry{
						dirEntry{name: "specialDir", isDir: true},
					},
				},
			},
			want: Model{
				cwd: "/special",
				files: []fs.DirEntry{
					dirEntry{name: "specialDir", isDir: true},
				},
				cursor:              position{},
				cachedDirSelections: map[string]string{"/special": "specialDir"},
				maxRows:             10,
			},
			want1Nil: true,
		},
		{
			name: "Load same directory when files deleted",
			fields: fields{
				cwd: "/special",
				files: []fs.DirEntry{
					dirEntry{name: "dir1", isDir: true},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 1},
				cachedDirSelections: map[string]string{"/special": "xx"},
				maxRows:             1,
			},
			args: args{
				msg: dirLoaded{
					path:  "/special",
					files: []fs.DirEntry{},
				},
			},
			want: Model{
				cwd:                 "/special",
				files:               []fs.DirEntry{},
				cursor:              position{},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			want1Nil: true,
		},
		{
			name: "Load same directory when previous selected files deleted",
			fields: fields{
				cwd: "/special",
				files: []fs.DirEntry{
					dirEntry{name: "dir1", isDir: true},
					dirEntry{name: "file2", isDir: false},
					dirEntry{name: "file3", isDir: false},
				},
				cursor:              position{c: 2},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			args: args{
				msg: dirLoaded{
					path: "/special",
					files: []fs.DirEntry{
						dirEntry{name: "dir1", isDir: true},
						dirEntry{name: "file2", isDir: false},
					},
				},
			},
			want: Model{
				cwd: "/special",
				files: []fs.DirEntry{
					dirEntry{name: "dir1", isDir: true},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{c: 1},
				cachedDirSelections: map[string]string{"/special": "file2"},
				maxRows:             1,
			},
			want1Nil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				cfg:                 tt.fields.cfg,
				cwd:                 tt.fields.cwd,
				files:               tt.fields.files,
				cursor:              tt.fields.cursor,
				cachedDirSelections: tt.fields.cachedDirSelections,
				maxRows:             tt.fields.maxRows,
				sb:                  tt.fields.sb,
				lastError:           tt.fields.lastError,
			}

			iface, got1 := m.Update(tt.args.msg)

			got := iface.(Model)
			got.clearAnimAt = time.Time{} // don't test timings

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Model.Update() got = %v, want %v", got, tt.want)
			}

			if (got1 == nil) != tt.want1Nil {
				t.Errorf("got1() = %v, want1Nil %v", got1, tt.want1Nil)
			}

			var got1Result tea.Msg
			if got1 != nil {
				got1Result = got1()
			}

			if !reflect.DeepEqual(got1Result, tt.want1Result) {
				t.Errorf("got1() got1Result = %v, want1 %v", got1Result, tt.want1Result)
			}
		})
	}
}
