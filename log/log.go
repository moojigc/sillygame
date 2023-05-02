package log

import "fmt"

type Logger struct {
	Verbose bool
}

func (l *Logger) Print(s string, a ...any) {
	if l.Verbose {
		fmt.Printf(s, a...)
	}
}

var Default *Logger

func init() { Default = &Logger{} }
