package controller

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/alx99/fly/cmd"
	"github.com/alx99/fly/config"
	"github.com/alx99/fly/logger"
	"github.com/alx99/fly/model"
	"github.com/alx99/fly/ui"
	"github.com/gdamore/tcell/v2"
)

const id = "CRL"

type controller struct {
	ui ui.UI
	m  model.Model

	cmdChan        chan cmd.Command
	commandBuffer  string
	kbs            config.KeyBindings
	msgWindowFocus bool
	shutdown       bool
	shutDown       *sync.WaitGroup
}

// Start starts the controller
func Start(ui ui.UI, m model.Model) {
	c := controller{ui: ui, m: m, kbs: config.GetAllKeyBindings(), cmdChan: make(chan cmd.Command, 10), shutDown: &sync.WaitGroup{}}
	logger.LogMessage(id, "Started", logger.DEBUG)

	go c.commandLoop()
	c.eventLoop()
}

func (c *controller) commandLoop() {
	c.shutDown.Add(1)
	for command := range c.cmdChan {
		switch t := command.(type) {
		case cmd.Cmd:
			switch command {
			case cmd.MoveUp:
				c.m.Navigate(model.Up)
			case cmd.MoveDown:
				c.m.Navigate(model.Down)
			case cmd.MoveLeft:
				c.m.Navigate(model.Left)
			case cmd.MoveRight:
				c.m.Navigate(model.Right)
			case cmd.MoveTop:
				c.m.Navigate(model.Top)
			case cmd.MoveBottom:
				c.m.Navigate(model.Bottom)
			case cmd.MarkSelection:
				c.m.MarkFile()
			case cmd.ToggleShowHidden:
				c.m.ToggleShowHidden()
			default:
				logger.LogMessage(id, "Cmd not implemented", logger.ERROR)
			}
		case cmd.BoolCommand:
			cfg := config.GetConfig()
			switch t.GetCommand() {
			case cmd.DirCandy:
				setBoolValue(&cfg.UI.DirCandy, t)
			case cmd.DrawBox:
				setBoolValue(&cfg.UI.Border, t)
			case cmd.IndentAll:
				setBoolValue(&cfg.UI.IndentAll, t)
			case cmd.IndentMarks:
				setBoolValue(&cfg.UI.IndentMarks, t)
			case cmd.Rainbow:
				setBoolValue(&cfg.UI.Rainbow, t)
			default:
				logger.LogMessage(id, "BoolCommand not implemented", logger.ERROR)
			}
			config.SetUIConfig(cfg.UI)
		default:
			logger.LogMessage(id, fmt.Sprintf("Hit default case. Nonexisten command: %+v", command), logger.ERROR)
		}
	}
	c.shutDown.Done()
}

func (c *controller) eventLoop() {
	for !c.shutdown {
		ev := c.ui.PollEvent()
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case *tcell.EventKey:
			if c.msgWindowFocus {
				c.handleKeyPressFocused(e)
			} else {
				c.handleKeyPressUnfocused(e)
			}
		case *tcell.EventInterrupt:
			logger.LogMessage(id, "eventinterrupt NOT IMPLEMENTED", logger.NORMAL)
		case *tcell.EventError:
			logger.LogError(id, "Received EventError", e)
		case *tcell.EventMouse:
			logger.LogMessage(id, "eventmouse NOT IMPLEMENTED", logger.NORMAL)
		default:
			logger.LogMessage(id, "Did not match on anything", logger.NORMAL)
		}
	}
	logger.LogMessage(id, "Shutting down", logger.DEBUG)
	logger.Shutdown()
}

func (c *controller) handleKeyPressUnfocused(e *tcell.EventKey) {
	if e.Key() == tcell.KeyRune {
		c.commandBuffer += string(e.Rune())
	} else {
		c.commandBuffer += e.Name()
	}

	mappings, ok := c.kbs.FindMatches(e)
	if !ok { // No keybindings found
		msg := "Sequence " + c.commandBuffer + " is unmapped"
		logger.LogError(id, msg, errors.New("No sequence found for keybinding"))
		c.ui.ShowMessage(ui.CreateMessage(msg, true))
		c.commandBuffer = ""
		c.kbs = config.GetAllKeyBindings()
		return
	}
	c.kbs = mappings

	if c.kbs.IsSingleKeyBinding() {
		c.commandBuffer = ""
		// Below are commands that need to be handled immediately
		// since they change how forfthcoming keypresses are interpreted
		switch c.kbs.GetCommand() {
		case cmd.Quit:
			close(c.cmdChan)
			c.shutDown.Wait()
			c.ui.Shutdown()
			c.shutdown = true
		case cmd.ToggleCommandMenu:
			if !c.msgWindowFocus {
				c.commandBuffer = ":"
				c.ui.ShowMessage(ui.CreateMessage(c.commandBuffer, false))
			}
			c.msgWindowFocus = !c.msgWindowFocus
		default:
			c.cmdChan <- c.kbs.GetCommand()
		}
		c.kbs = config.GetAllKeyBindings()
	}
}

func (c *controller) handleKeyPressFocused(e *tcell.EventKey) {
	tK := e.Key()
	switch {
	case tK == tcell.KeyEsc:
		c.ui.CloseMsgWindow()
		c.msgWindowFocus = !c.msgWindowFocus
		c.commandBuffer = ""
	case tK == tcell.KeyBackspace2 || tK == tcell.KeyBackspace:
		c.commandBuffer = c.commandBuffer[:len(c.commandBuffer)-1]
		c.ui.ShowMessage(ui.CreateMessage(c.commandBuffer, false))
		if c.commandBuffer == "" {
			c.ui.CloseMsgWindow()
			c.msgWindowFocus = false
		}
	case tK == tcell.KeyEnter:
		cmd, err := parseCommand(c.commandBuffer[1:])
		if err != nil {
			c.ui.ShowMessage(ui.CreateMessage(err.Error(), true))
			logger.LogError(id, "Error while parsing command", err)
		} else {
			c.cmdChan <- cmd
		}

		c.ui.CloseMsgWindow()
		c.msgWindowFocus = !c.msgWindowFocus
		c.commandBuffer = ""
	default:
		// A non printable character was sent
		if e.Key() != tcell.KeyRune {
			logger.LogError(id, "Received "+e.Name(), errors.New("Received non printable character in the msgWindow"))
		} else {
			c.commandBuffer += string(e.Rune())
			c.ui.ShowMessage(ui.CreateMessage(c.commandBuffer, false))
		}
	}
}
func parseCommand(command string) (cmd.Command, error) {
	s := strings.Split(command, " ")
	if s[0] == "toggle" {
		if len(s) < 2 {
			return nil, errors.New("too few arguments")
		}
		if c, ok := cmd.ParseCommand(s[1]); ok {
			return cmd.CreateBoolCommand(c), nil
		}
		return nil, errors.New("command '" + s[1] + "' not found!")
	}
	return nil, errors.New("command '" + command + "' not found")
}

// helper to set a value from a CommandBoolean
func setBoolValue(b *bool, c cmd.BoolCommand) {
	if c.HasValueSet() {
		*b = c.GetValue()
	}
	// If a BoolCommand has no value set
	// is interpreted as a toggle
	*b = !*b
}
