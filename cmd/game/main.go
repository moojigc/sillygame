package main

import (
	"flag"
	"math"

	"github.com/moojigc/routines_game/app"
	"github.com/moojigc/routines_game/controller"
)

var gameOptions = &app.GameOptions{}

func main() {
	flag.BoolVar(&gameOptions.LogRounds, "verbose", false, "Log each round?")
	flag.IntVar(&gameOptions.PlayerCount, "players", 2, "How many players?")
	flag.Int64Var(&gameOptions.Rounds, "rounds", math.MaxInt64, "How many rounds?")

	flag.Parse()

	app.Run(gameOptions, controller.New())
}
