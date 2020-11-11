package ui

import (
	"github.com/alx99/fly/config"
	"github.com/alx99/fly/logger"
	"github.com/alx99/fly/model"
	"github.com/alx99/fly/ui/pos"
	"github.com/gdamore/tcell/v2"
)

// UI log identifier
const id = "TUI"

// UI interface exposed to the controller
type UI interface {
	Shutdown()
	// todo the Sync and Refresh method should not be exposed, the UI should always know when it needs to re-render itself without anything explicitly telling it to
	Sync()
	Refresh()
	PollEvent() tcell.Event
	Resize()
	CloseMsgWindow()
	ShowMessage(string)
}

// todo the tui should not really need access to the log
// ? or maybe it should
type ui struct {
	screen           tcell.Screen
	wd, pd, cd       *FileWindow
	w, h             int
	msgWindowVisible bool
	mw               *msgWindow

	m   model.Model
	cfg config.UI
}

func (ui *ui) start() (UI, error) {
	logger.LogMessage(id, "Starting", logger.DEBUG)
	s, err := tcell.NewScreen()
	if err != nil {
		logger.LogError(id, "Tried to create a new screen", err)
		return nil, err
	}
	if err := s.Init(); err != nil {
		logger.LogError(id, "Tried to initiate a new screen", err)
		return nil, err
	}

	ui.screen = s

	ui.initWindows()
	return ui, nil
}

func (ui *ui) initWindows() {
	// todo save space mode
	// render each window with exactly the width it needs
	// if it needs more than what it normal is assigned then
	// cut the filename paths

	tmpCoord := pos.NewCoord(0, 0)
	ui.pd = CreateFileWindow(pos.CreateArea(tmpCoord, tmpCoord, pos.CreatePadding(0, 0, 0, 0)), ui.screen)
	ui.wd = CreateFileWindow(pos.CreateArea(tmpCoord, tmpCoord, pos.CreatePadding(1, 0, 0, 0)), ui.screen)
	ui.cd = CreateFileWindow(pos.CreateArea(tmpCoord, tmpCoord, pos.CreatePadding(1, 0, 0, 0)), ui.screen)
	ui.mw = createMsgWindow(pos.CreateArea(tmpCoord, tmpCoord, pos.CreatePadding(0, 0, 0, 0)), ui.screen)
}

func (ui ui) PollEvent() tcell.Event {
	return ui.screen.PollEvent()
}

// todo should not be exported
// Sync displays the new changes
func (ui ui) Sync() {
	// todo in the future every window should be responsible for clearing itself, and redrawing borders and stuff all the time won't be necessary
	ui.screen.Clear()

	if ui.cfg.Border {
		if ui.msgWindowVisible {
			// Border around everything except the last row of the screen
			drawOutline(pos.NewCoord(0, 0), pos.NewCoord(ui.w, ui.h-1), ui.screen, tcell.StyleDefault)
			ui.screen.SetContent(0, ui.wd.a.GetEnd().Y+1, tcell.RuneLLCorner, nil, tcell.StyleDefault)
			ui.screen.SetContent(ui.w, ui.wd.a.GetEnd().Y+1, tcell.RuneLRCorner, nil, tcell.StyleDefault)
		} else {
			// Border around everything
			drawOutline(pos.NewCoord(0, 0), pos.NewCoord(ui.w, ui.h), ui.screen, tcell.StyleDefault)
		}
	}

	ui.pd.RenderDir(ui.m.GetPD(), ui.cfg)
	ui.wd.RenderDir(ui.m.GetWD(), ui.cfg)
	ui.cd.RenderDir(ui.m.GetCD(), ui.cfg)

	if ui.msgWindowVisible {
		ui.mw.show()
	}
	ui.screen.Show()
}

// Refresh refreshes every single cell, currently it is not used
func (ui ui) Refresh() {
	ui.screen.Sync()
}

func (ui ui) Shutdown() {
	logger.LogMessage(id, "Shutting down", logger.DEBUG)
	ui.screen.Clear()
	ui.screen.Fini()
}

func (ui *ui) Resize() {
	w, h := ui.screen.Size()
	// These last x and y can't be rendered on
	w--
	h--
	// Set new size
	ui.w = w
	ui.h = h

	// Calculate the position of everything
	baseRatio := float64(w) / float64(ui.cfg.PDRatio+ui.cfg.WDRatio+ui.cfg.CDRatio)
	pdWidth := int(baseRatio * ui.cfg.PDRatio)
	yStart := 0
	xStart := 0
	wdStart := pdWidth + 1
	cdStart := pdWidth + int(baseRatio*ui.cfg.WDRatio) + 1

	if ui.msgWindowVisible {
		ui.mw.SetPos(pos.NewCoord(0, h), pos.NewCoord(w, h))
		// Squish height by 1px, which it taken up by the messageWindow
		h--
	}

	// If we're displaying a border wd and cd are squished by 1px in x direction
	if ui.cfg.Border {
		h--
		w--
		xStart++
		yStart++
		wdStart++
	}

	// todo need to write some kind of test for this
	/*
		ui.lgr.LogMessage(id, "pd render from "+strconv.Itoa(pdStart)+" to "+strconv.Itoa(wdStart-1), logger.DEBUG)
		ui.lgr.LogMessage(id, "wd render from "+strconv.Itoa(wdStart)+" to "+strconv.Itoa(cdStart-1), logger.DEBUG)
		ui.lgr.LogMessage(id, "cd render from "+strconv.Itoa(cdStart)+" to "+strconv.Itoa(w), logger.DEBUG)
	*/

	// Update positions of filewindows
	ui.pd.SetPos(pos.NewCoord(xStart, yStart), pos.NewCoord(wdStart-1, h))
	ui.wd.SetPos(pos.NewCoord(wdStart, yStart), pos.NewCoord(cdStart-1, h))
	ui.cd.SetPos(pos.NewCoord(cdStart, yStart), pos.NewCoord(w, h))

}

func (ui *ui) CloseMsgWindow() {
	ui.msgWindowVisible = false
	// We have to resize here since we have
	// to recalculate where filewindows start
	// and end
	ui.Resize()
}
func (ui *ui) ShowMessage(msg string) {
	if !ui.msgWindowVisible {
		ui.msgWindowVisible = true
		ui.Resize()
	}
	ui.mw.setMessage(msg)
}

func (ui *ui) uiChange(cfg config.UI) {
	ui.cfg = cfg
	ui.Resize()
}

// Start starts up the UI
func Start(m model.Model) (UI, error) {
	ui := ui{m: m, cfg: config.GetConfig().UI}
	config.AttachConfigObserver(ui.uiChange)
	return ui.start()
}

// drawOutline draws a box in an area
func drawOutline(start, end pos.Coord, s tcell.Screen, st tcell.Style) {
	x1, x2 := start.X, end.X
	y1, y2 := start.Y, end.Y

	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	// Draw borders
	for col := x1; col <= x2; col++ {
		s.SetContent(col, y1, tcell.RuneHLine, nil, st)
		s.SetContent(col, y2, tcell.RuneHLine, nil, st)
	}
	for row := y1 + 1; row < y2; row++ {
		s.SetContent(x1, row, tcell.RuneVLine, nil, st)
		s.SetContent(x2, row, tcell.RuneVLine, nil, st)
	}

	// Only draw corners if necessary
	if y1 != y2 && x1 != x2 {
		s.SetContent(x1, y1, tcell.RuneULCorner, nil, st)
		s.SetContent(x2, y1, tcell.RuneURCorner, nil, st)
		s.SetContent(x1, y2, tcell.RuneLLCorner, nil, st)
		s.SetContent(x2, y2, tcell.RuneLRCorner, nil, st)
	}
}
