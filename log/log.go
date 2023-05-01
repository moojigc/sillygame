package log

import "fmt"

type Logger struct {
	Include bool
}

func (l *Logger) Print(s string, a ...any) {
	if l.Include {
		fmt.Printf(s, a...)
	}
}

var Default *Logger

func init() { Default = &Logger{} }
