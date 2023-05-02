package controller

import (
	"fmt"
)

type Controller struct{}

func (c Controller) Announce(announcement string, a ...any) {
	fmt.Printf(announcement+"\n", a...)
}
func (c Controller) AskInt(question string, answer *int) {
	c.Announce(question)
	var input int
	fmt.Scanln(&input)
	*answer = input
}
func (c Controller) AskString(question string, answer *string) {
	c.Announce(question)
	var input string
	fmt.Scanln(&input)
	*answer = input
}

func New() Controller {
	return Controller{}
}
