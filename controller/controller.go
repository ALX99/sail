package controller

import (
	"errors"
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
	uiMessageChan  chan<- ui.Message
	commandBuffer  string
	msgWindowFocus bool
	shutDown       *sync.WaitGroup
}

// Start starts the controller
func Start(ui ui.UI, m model.Model) {
	c := controller{ui: ui, m: m, cmdChan: make(chan cmd.Command, 10), shutDown: &sync.WaitGroup{}}
	logger.LogMessage(id, "Started", logger.DEBUG)
	c.uiMessageChan = ui.GetMessageChan()

	go c.commandLoop()
	c.eventLoop()
}

func (c *controller) commandLoop() {
	c.shutDown.Add(1)
	for command := range c.cmdChan {
		switch t := command.(type) {
		case cmd.Cmd:
			switch command.GetCommand() {
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
			default:
				logger.LogMessage(id, "Not implemented", logger.DEBUG)
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
			}
			config.SetUIConfig(cfg.UI)
		}
	}
	c.shutDown.Done()
}
func (c *controller) eventLoop() {
	var m map[config.Key]config.KeyBinding
	var command cmd.Command

loop:
	for {
		ev := c.ui.PollEvent()
		if ev == nil {
			continue
		}
		switch e := ev.(type) {
		case *tcell.EventKey:
			k := config.EventKeyToKey(e)
			if !c.msgWindowFocus {
				command, m = config.MatchCommand(k, m)

				// Here we actually found a keybinding
				if command != nil {
					switch command.GetCommand() {
					case cmd.Quit:
						close(c.cmdChan)
						c.shutDown.Wait()
						c.ui.Shutdown()
						break loop
					case cmd.ToggleCommandMenu:
						// The reason this is toggled here is to avoid subsequent keypresses after a ToggleCommandMenu command being interpreted as a keybinding instead of the key going to the commandbuffer.
						// We need to do this since the commandLoop runs in another goroutine
						if !c.msgWindowFocus {
							c.commandBuffer = ":"
							c.uiMessageChan <- ui.CreateMessage(c.commandBuffer, false)
						}
						c.msgWindowFocus = !c.msgWindowFocus
					default:
						c.cmdChan <- command
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
					c.uiMessageChan <- ui.CreateMessage(c.commandBuffer, false)
					if c.commandBuffer == "" {
						c.ui.CloseMsgWindow()
						c.msgWindowFocus = false
					}
				case tK == tcell.KeyEnter:
					cmd, err := parseCommand(c.commandBuffer[1:])
					if err != nil {
						c.uiMessageChan <- ui.CreateMessage(err.Error(), true)
					} else {
						c.cmdChan <- cmd
					}
					c.ui.CloseMsgWindow()
					c.msgWindowFocus = !c.msgWindowFocus
					c.commandBuffer = ""
				default:
					c.commandBuffer += k.String()
					c.uiMessageChan <- ui.CreateMessage(c.commandBuffer, false)
				}
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

func parseCommand(command string) (cmd.Command, error) {
	s := strings.Split(command, " ")
	if s[0] == "toggle" {
		if len(s) < 2 {
			return nil, errors.New("Too few arguments")
		}
		if c, ok := cmd.ParseCommand(s[1]); ok {
			return cmd.CreateBoolCommand(c), nil
		}
		return nil, errors.New("Command '" + s[1] + "' not found!")
	}
	return nil, nil
}

// helper to set a value from a CommandBoolean
func setBoolValue(b *bool, c cmd.BoolCommand) {
	if c.HasValueSet() {
		*b = c.GetValue()
	}
	// If a BoolCommand has no value set it
	// is interpreted as a toggle
	*b = !*b
}
