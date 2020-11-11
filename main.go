package main

import (
	"fmt"
	"log"

	"github.com/alx99/fly/controller"
	"github.com/alx99/fly/logger"
	"github.com/alx99/fly/model"
	"github.com/alx99/fly/ui"
)

// todo spawn a shell inside the terminal. just open a big box with terminal access
func main() {
	if err := logger.Start("fly.log", logger.ERROR); err != nil {
		log.Fatalln(err)
	}
	m, err := model.CreateModel()
	if err != nil {
		log.Fatalln(err)
	}
	ui, err := ui.Start(m)
	if err != nil {
		fmt.Println("Something went wrong. Read the Logfile")
		return
	}
	controller.Start(ui, m)
}
