package controller

import (
	"strings"
	"sync"

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

	cmdChan        chan config.Command
	commandBuffer  string
	msgWindowFocus bool
	shutDown       *sync.WaitGroup
}

// Start starts the controller
func Start(ui ui.UI, m model.Model) {
	c := controller{ui: ui, m: m, cmdChan: make(chan config.Command, 10), shutDown: &sync.WaitGroup{}}
	logger.LogMessage(id, "Started", logger.DEBUG)

	go c.commandLoop()
	c.eventLoop()
}

func (c *controller) commandLoop() {
	c.shutDown.Add(1)
	for cmd := range c.cmdChan {
		switch cmd {
		case config.MoveUp:
			c.m.Navigate(model.Up)
		case config.MoveDown:
			c.m.Navigate(model.Down)
		case config.MoveLeft:
			c.m.Navigate(model.Left)
		case config.MoveRight:
			c.m.Navigate(model.Right)
		case config.MoveTop:
			c.m.Navigate(model.Top)
		case config.MoveBottom:
			c.m.Navigate(model.Bottom)
		case config.MarkSelection:
			c.m.MarkFile()
		case config.DirCandy:
			cfg := config.GetConfig()
			cfg.UI.DirCandy = !cfg.UI.DirCandy
			config.SetUIConfig(cfg.UI)
		case config.DrawBox:
			cfg := config.GetConfig()
			cfg.UI.Border = !cfg.UI.Border
			config.SetUIConfig(cfg.UI)
		case config.IndentAll:
			cfg := config.GetConfig()
			cfg.UI.IndentAll = !cfg.UI.IndentAll
			config.SetUIConfig(cfg.UI)
		case config.IndentMarks:
			cfg := config.GetConfig()
			cfg.UI.IndentMarks = !cfg.UI.IndentMarks
			config.SetUIConfig(cfg.UI)
		case config.Rainbow:
			cfg := config.GetConfig()
			cfg.UI.Rainbow = !cfg.UI.Rainbow
			config.SetUIConfig(cfg.UI)
		case config.Nil:
			// No keybinding defined here
			logger.LogMessage(id, "Not defined", logger.DEBUG)
			continue
		default:
			logger.LogMessage(id, "Not implemented", logger.DEBUG)
			continue
		}
		c.ui.Sync()
	}
	c.shutDown.Done()
}
func (c *controller) eventLoop() {
	var m map[config.Key]config.KeyBinding
	var cmd config.Command

loop:
	for {
		switch e := c.ui.PollEvent().(type) {
		case *tcell.EventResize:
			c.ui.Resize()
			c.ui.Sync()

		case *tcell.EventKey:
			k := config.EventKeyToKey(e)
			if !c.msgWindowFocus {
				cmd, m = config.MatchCommand(k, m)
				if m == nil {
					if cmd == config.Quit {
						close(c.cmdChan)
						c.shutDown.Wait()
						c.ui.Shutdown()
						break loop
					} else if cmd != config.OpenCommandMenu {
						c.cmdChan <- cmd
					} else {
						// The reason this is toggled here is to avoid subsequent keypresses after a ToggleCommandMenu command being interpreted as a keybinding instead of the key going to the commandbuffer.
						// We need to do this since the commandLoop runs in another goroutine
						if !c.msgWindowFocus {
							c.commandBuffer = ":"
							c.ui.ShowMessage(c.commandBuffer)
							c.ui.Sync()
						}
						c.msgWindowFocus = !c.msgWindowFocus
					}
				}
			} else {
				tK := e.Key()
				switch {
				case tK == tcell.KeyEsc:
					c.ui.CloseMsgWindow()
					c.msgWindowFocus = !c.msgWindowFocus
					c.commandBuffer = ""
				case tK == tcell.KeyBackspace2 || tK == tcell.KeyBackspace:
					c.commandBuffer = c.commandBuffer[:len(c.commandBuffer)-1]
					c.ui.ShowMessage(c.commandBuffer)
					if c.commandBuffer == "" {
						c.ui.CloseMsgWindow()
						c.msgWindowFocus = false
					}
				case tK == tcell.KeyEnter:
					cmd, err := c.parseCommand()
					if err != nil {
						// todo show error
					} else {
						c.cmdChan <- cmd
					}
					c.ui.CloseMsgWindow()
					c.msgWindowFocus = !c.msgWindowFocus
					c.commandBuffer = ""
				default:
					c.commandBuffer += k.String()
					c.ui.ShowMessage(c.commandBuffer)
				}
				c.ui.Sync()
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

func (c controller) parseCommand() (config.Command, error) {
	s := strings.Split(c.commandBuffer[1:], " ")
	if s[0] == "toggle" {
		if len(s) < 2 {
			// todo error
			return config.Nil, nil
		}
		if c, ok := config.ParseCommand(s[1]); ok {
			return c, nil
		}
		// todo error

	}
	return config.Nil, nil
}
