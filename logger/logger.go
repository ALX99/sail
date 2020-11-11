package logger

import (
	"log"
	"os"
	"sync"
)

// Log identifier
const id = "LOG"

var lgr logger

type logMessage struct {
	sender, message string
	level           LogLevel
}

type logger struct {
	file     *os.File
	level    LogLevel
	log      *log.Logger
	mChan    chan logMessage
	quitSync *sync.WaitGroup
}

// LogMessage logs a message to the log
func LogMessage(sender, message string, level LogLevel) {
	lgr.mChan <- logMessage{sender: sender, message: message, level: level}
}

// LogError logs an error to the log
func LogError(sender, message string, err error) {
	lgr.mChan <- logMessage{sender: sender, message: message + ". error=" + err.Error(), level: ERROR}
}

// Shutdown shuts down the logger
func Shutdown() {
	LogMessage(id, "Shutting down", DEBUG)
	close(lgr.mChan)
	lgr.quitSync.Wait()
}

func (l logger) logMessages() {
	l.quitSync.Add(1)
	for m := range l.mChan {
		if l.level < m.level {
			continue
		}
		l.log.Println(m.sender + ": " + m.message)
	}
	l.file.Close()
	l.quitSync.Done()
}

// Start starts the logger
func Start(filename string, level LogLevel) error {
	f, err := os.OpenFile(filename,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// Create a new logger
	l := logger{level: level, mChan: make(chan logMessage, 1000), quitSync: &sync.WaitGroup{}}
	l.log = log.New(f, "", log.Ltime)
	l.file = f

	lgr = l
	go lgr.logMessages()
	LogMessage(id, "Starting", DEBUG)
	return nil
}
