package models

import (
	"errors"
	"io/fs"
	"reflect"
	"slices"
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
	type mocks struct {
		removeFile func(Model) func(string) error
		readDir    func(Model) func(string) ([]fs.DirEntry, error)
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		want       Model
		wantMsgs   []tea.Msg
		mocks      mocks
		filterMsgs []tea.Msg // msgs not to resend to model
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
			filterMsgs: []tea.Msg{clearPrevCWD{}},
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
			filterMsgs: []tea.Msg{clearPrevCWD{}},
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
			filterMsgs: []tea.Msg{clearPrevCWD{}},
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
		},
		{
			name: "Delete last file in column",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{Delete: "d"}},
				},
				cwd: "/test",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             1,
			},
			mocks: mocks{
				removeFile: func(m Model) func(string) error {
					return func(string) error { return nil }
				},
				readDir: func(m Model) func(string) ([]fs.DirEntry, error) {
					return func(string) ([]fs.DirEntry, error) {
						return slices.DeleteFunc(m.files, func(f fs.DirEntry) bool {
							return f.Name() == "file2"
						}), nil
					}
				},
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{Delete: "d"}},
				},
				cwd: "/test",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
				},
				cursor:              position{r: 0, c: 0},
				cachedDirSelections: map[string]string{"/test": "file1"},
				maxRows:             1,
			},
			wantMsgs: []tea.Msg{
				dirLoaded{
					path:  "/test",
					files: []fs.DirEntry{dirEntry{name: "file1", isDir: false}},
				},
			},
		},
		{
			name: "Delete last file in row",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{Delete: "d"}},
				},
				cwd: "/test",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
					dirEntry{name: "file3", isDir: false},
				},
				cursor:              position{r: 2, c: 0},
				cachedDirSelections: map[string]string{},
				maxRows:             3,
			},
			mocks: mocks{
				removeFile: func(m Model) func(string) error {
					return func(string) error { return nil }
				},
				readDir: func(m Model) func(string) ([]fs.DirEntry, error) {
					return func(string) ([]fs.DirEntry, error) {
						return []fs.DirEntry{
							dirEntry{name: "file1", isDir: false},
							dirEntry{name: "file2", isDir: false},
						}, nil
					}
				},
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			want: Model{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{Delete: "d"}},
				},
				cwd: "/test",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{r: 1, c: 0},
				cachedDirSelections: map[string]string{"/test": "file2"},
				maxRows:             3,
			},
			wantMsgs: []tea.Msg{
				dirLoaded{
					path: "/test",
					files: []fs.DirEntry{
						dirEntry{name: "file1", isDir: false},
						dirEntry{name: "file2", isDir: false},
					},
				},
			},
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

			if tt.mocks.removeFile != nil {
				removeFile = tt.mocks.removeFile(m)
				readDir = tt.mocks.readDir(m)
			}

			var cmd tea.Cmd
			var gotMsgs []tea.Msg
			var iface tea.Model = m
			msg := tt.args.msg
			for {
				iface, cmd = iface.Update(msg)
				if cmd == nil {
					break
				}

				msg = cmd()
				if msg == nil {
					break
				}

				if shouldIgnoreMsg(msg, tt.filterMsgs) {
					break
				}
				gotMsgs = append(gotMsgs, msg)
			}

			for i, msg := range gotMsgs {
				if len(tt.wantMsgs) <= i {
					t.Errorf("Model.Update() gotMsgs = %v, wantMsgs %v", gotMsgs, tt.wantMsgs)
				} else if !reflect.DeepEqual(msg, tt.wantMsgs[i]) {
					t.Errorf("Model.Update() gotMsgs = %v, wantMsgs %v", gotMsgs, tt.wantMsgs)
				}
			}

			got := iface.(Model)
			got.clearAnimAt = time.Time{} // don't test timings

			if got.cursor.r != tt.want.cursor.r || got.cursor.c != tt.want.cursor.c {
				t.Errorf("Model.Update() got cursor = %+v, want cursor %+v", got.cursor, tt.want.cursor)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Model.Update() got, want:\n%v\n%v\n", got, tt.want)
			}
		})
	}
}

func shouldIgnoreMsg(msg tea.Msg, filterMsgs []tea.Msg) bool {
	for _, fMsg := range filterMsgs {
		if reflect.DeepEqual(fMsg, msg) {
			return true
		}
	}
	return false
}
