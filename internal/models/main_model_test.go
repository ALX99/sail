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
		fs fsys
	}

	tests := []struct {
		name       string
		fields     fields
		args       args
		wantFunc   func(Model) Model
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
			wantFunc: func(m Model) Model {
				m.cwd = "/go/src"
				m.files = []fs.DirEntry{
					dirEntry{name: "goa", isDir: true},
					dirEntry{name: "go", isDir: true},
				}
				m.prevCWD = "/go"
				return m
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
			wantFunc: func(m Model) Model {
				m.cwd = "/cache"
				m.files = []fs.DirEntry{
					dirEntry{name: "file1", isDir: true},
					dirEntry{name: "ex", isDir: true},
				}
				m.prevCWD = "/currpath"
				m.cursor = position{r: 1, c: 0}
				return m
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
			wantFunc: func(m Model) Model {
				m.cwd = "/testpath/dir1"
				m.files = []fs.DirEntry{}
				m.cursor = position{}
				m.cachedDirSelections = map[string]string{"/testpath": "dir1"}
				m.prevCWD = "/testpath"
				return m
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
				cursor:              position{r: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			},
			wantFunc: func(m Model) Model {
				m.cursor = position{}
				return m
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
			wantFunc: func(m Model) Model {
				return m
			},
		},
		{
			name: "Test NavUp when cursor is at the top of the current column",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{NavUp: "a"}},
				},
				cwd: "/testpath",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
				},
				cursor:              position{c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			},
			wantFunc: func(m Model) Model {
				return m
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
			wantFunc: func(m Model) Model {
				m.cursor = position{r: 1, c: 0}
				return m
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
				cursor:              position{r: 1, c: 0},
				cachedDirSelections: map[string]string{},
				maxRows:             10,
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			},
			wantFunc: func(m Model) Model {
				return m
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
			wantFunc: func(m Model) Model {
				return m
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
			wantFunc: func(m Model) Model {
				m.cursor = position{r: 0, c: 0}
				return m
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
			wantFunc: func(m Model) Model {
				m.cursor = position{r: 0, c: 1}
				return m
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
			wantFunc: func(m Model) Model {
				return m
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
				lastError:           errors.New("some error"),
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}},
			},
			wantFunc: func(m Model) Model {
				m.lastError = nil
				return m
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
			wantFunc: func(m Model) Model {
				m.files = []fs.DirEntry{
					dirEntry{name: "specialDir", isDir: true},
				}
				return m
			},
		},
		{
			name: "Load same directory when files deleted",
			fields: fields{
				cwd: "/special",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
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
			wantFunc: func(m Model) Model {
				m.files = []fs.DirEntry{}
				m.cursor = position{}
				return m
			},
		},
		{
			name: "Load same directory when previous selected files deleted",
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
			wantFunc: func(m Model) Model {
				m.files = []fs.DirEntry{}
				m.cursor = position{}
				return m
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
				fs: mockOS{},
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			wantFunc: func(m Model) Model {
				m.cursor = position{}
				m.files = []fs.DirEntry{dirEntry{name: "file1", isDir: false}}

				return m
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
				fs: mockOS{},
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			wantFunc: func(m Model) Model {
				m.cursor = position{r: 1, c: 0}
				m.files = slices.Delete(m.files, 2, 3)
				return m
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
					dirEntry{name: "file3", isDir: false},
				},
				cursor:              position{r: 0, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             2,
			},
			mocks: mocks{
				fs: mockOS{},
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			wantFunc: func(m Model) Model {
				m.cursor = position{r: 1, c: 0}
				m.files = slices.Delete(m.files, 2, 3)
				return m
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
		{
			name: "Delete file in the middle",
			fields: fields{
				cfg: config.Config{
					Settings: config.Settings{Keymap: config.Keymap{Delete: "d"}},
				},
				cwd: "/test",
				files: []fs.DirEntry{
					dirEntry{name: "file1", isDir: false},
					dirEntry{name: "file2", isDir: false},
					dirEntry{name: "file3", isDir: false},
					dirEntry{name: "file4", isDir: false},
					dirEntry{name: "file5", isDir: false},
					dirEntry{name: "file6", isDir: false},
					dirEntry{name: "file7", isDir: false},
					dirEntry{name: "file8", isDir: false},
					dirEntry{name: "file9", isDir: false},
				},
				cursor:              position{r: 1, c: 1},
				cachedDirSelections: map[string]string{},
				maxRows:             3,
			},
			mocks: mocks{
				fs: mockOS{},
			},
			args: args{
				msg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			},
			wantFunc: func(m Model) Model {
				m.cursor = position{r: 1, c: 1}
				m.files = slices.Delete(m.files, 4, 5)
				return m
			},
			wantMsgs: []tea.Msg{
				dirLoaded{
					path: "/test",
					files: []fs.DirEntry{
						dirEntry{name: "file1", isDir: false},
						dirEntry{name: "file2", isDir: false},
						dirEntry{name: "file3", isDir: false},
						dirEntry{name: "file4", isDir: false},
						dirEntry{name: "file6", isDir: false},
						dirEntry{name: "file7", isDir: false},
						dirEntry{name: "file8", isDir: false},
						dirEntry{name: "file9", isDir: false},
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

			if mock, ok := tt.mocks.fs.(mockOS); ok {
				mock.addFromModel(m)
				osi = mock
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

			want := tt.wantFunc(m)

			if got.cursor.r != want.cursor.r || got.cursor.c != want.cursor.c {
				t.Errorf("Model.Update() got cursor = %+v, want cursor %+v", got.cursor, want.cursor)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Model.Update() got, want:\n%v\n%v\n", got, want)
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
