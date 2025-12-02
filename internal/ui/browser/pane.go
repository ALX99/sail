package browser

import (
	"github.com/alx99/sail/internal/filesys"
	"github.com/alx99/sail/internal/ui/components/filelist"
	"golang.org/x/text/collate"
)

type pane struct {
	view  *filelist.View
	cache map[string]filelist.State
}

func newPane(path string, state filelist.State, coll *collate.Collator, checker filelist.SelChecker, highlight bool) *pane {
	return &pane{
		view:  filelist.New(path, state, checker, coll, highlight),
		cache: make(map[string]filelist.State, 32),
	}
}

func (p *pane) View() string {
	return p.view.View()
}

func (p *pane) Path() string {
	return p.view.Path()
}

func (p *pane) SetBounds(rows, cols int) {
	p.view.SetMaxDims(rows, cols)
}

func (p *pane) MoveUp() {
	p.view.MoveUp()
}

func (p *pane) MoveDown() {
	p.view.MoveDown()
}

func (p *pane) SelectedRow() int {
	return p.view.SelectedRow()
}

func (p *pane) CurrEntry() (filesys.DirEntry, bool) {
	return p.view.CurrEntry()
}

func (p *pane) RememberCurrent() {
	if p.view.Path() == "" {
		return
	}
	p.cache[p.view.Path()] = p.view.State()
}

func (p *pane) SetDir(dir filesys.Dir, override filelist.State) {
	state := p.cache[dir.Path()]
	if override.SelectedName != "" {
		state.SelectedName = override.SelectedName
	}
	if override.ViewportStart != 0 {
		state.ViewportStart = override.ViewportStart
	}

	p.view.ChDir(dir, state)
	p.cache[dir.Path()] = p.view.State()
}

func (p *pane) SetShowHidden(show bool) {
	p.view.SetShowHidden(show)
}
