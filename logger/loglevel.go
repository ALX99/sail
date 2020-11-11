package logger

// LogLevel is the level of the log message
type LogLevel int

// Different type of loglevels
const (
	DEBUG LogLevel = iota
	NORMAL
	ERROR
)

func (l LogLevel) String() string {
	if l == DEBUG {
		return "Debug"
	} else if l == NORMAL {
		return "Normal"
	} else {
		return "Error"
	}
}
