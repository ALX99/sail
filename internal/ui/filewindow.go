package ui

import (
	"io/fs"
	"os"
	"strings"

	"github.com/alx99/fly/internal/util"
	tea "github.com/charmbracelet/bubbletea"
)

type Direction uint8
type ID uint8

const (
	Up Direction = iota
	Down
)

type fileWindowMsg struct {
	msg interface{}
	to  ID
}

type fileWindow struct {
	path     string
	files    []fs.DirEntry
	moveDown bool

	id             ID
	h, w           int
	pos            int
	prevFileStart  int
	fileStart      int
	visibleFileLen int

	// Configurable settings
	scrollPadding int
}

func NewFileWindow(path string, id ID) fileWindow {
	return fileWindow{id: id, path: path, scrollPadding: 2}
}

// Init will return the necessary tea.Msg for the fileWindow
// to become initialized and ready
func (fw fileWindow) Init() tea.Msg {
	files, err := os.ReadDir(fw.path)
	if err != nil {
		panic(err) // todo
	}
	return fileWindowMsg{to: fw.id, msg: files}
}

func (fw fileWindow) Update(msg tea.Msg) (fileWindow, tea.Cmd) {
	switch msg := msg.(type) {
	case fileWindowMsg:
		if msg.to != fw.id {
			break
		}
		fw.files = msg.msg.([]os.DirEntry)
		fw.visibleFileLen = len(fw.files)

	case tea.WindowSizeMsg:
		fw.h, fw.w = msg.Height, msg.Width
	}

	return fw, nil
}

func (fw fileWindow) View() string {
	var nameBuilder strings.Builder
	names := make([]string, 0, fw.visibleFileLen)
	drawn := 0

	for i := fw.fileStart; i < fw.visibleFileLen && drawn < fw.h; i++ {
		drawn++
		if i == fw.pos {
			nameBuilder.WriteString("> ")
		}

		nameBuilder.WriteString(fw.files[i].Name())

		if fw.files[i].IsDir() {
			nameBuilder.WriteString("/")
		}

		names = append(names, nameBuilder.String())
		nameBuilder.Reset()
	}

	return strings.Join(names, "\n")
}

func (fw *fileWindow) move(dir Direction) {
	if dir == Up {
		if fw.pos > 0 {
			fw.pos -= 1

			if fw.fileStart > (fw.pos - fw.scrollPadding) {
				fw.fileStart = util.Max(0, fw.fileStart-1)
			}
		}
	} else {
		if fw.pos < fw.visibleFileLen-1 {
			fw.pos += 1

			if fw.pos-fw.fileStart+1 > (fw.h - fw.scrollPadding) {
				fw.fileStart = util.Min(fw.visibleFileLen-fw.h, fw.fileStart+1)
			}
		}
	}

}
