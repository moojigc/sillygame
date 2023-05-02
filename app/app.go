package app

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	col "github.com/moojigc/routines_game/collection"
	"github.com/moojigc/routines_game/controller"
	log "github.com/moojigc/routines_game/log"
	"github.com/moojigc/routines_game/scoretracker"
)

type Player struct {
	Name     string
	Location string
	Age      int
}

// randInt returns int64 in range [0, x)
func randInt(x int64) int64 {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(x))
	return nBig.Int64()
}

func routine(c *col.Collection, tracker *scoretracker.ScoreTracker, p *Player) {
	time.Sleep(time.Microsecond * time.Duration(randInt(10)))
	c.Add("name", p.Name).
		Add("location", p.Location).
		Add("age", p.Age)

	if !c.Has("is_a_winner") {
		c.Add("is_a_winner", true)
		tracker.Increment(p.Name)
	}
}

func randomizeSlice[T any](slice []T) {
	sliceLen := len(slice)
	for i := 0; i < sliceLen; i++ {
		from, to := randInt(int64(sliceLen)), randInt(int64(sliceLen))
		originalTo := slice[to]
		slice[to] = slice[from]
		slice[from] = originalTo
	}
}

func runOneRound(players []*Player, tracker *scoretracker.ScoreTracker, round int64) {
	log.Default.Print("Round %d\n", round)

	var wg sync.WaitGroup

	collection := col.New()

	randomizeSlice(players)
	log.Default.Print("Players are%s:\n", players)

	for i := 0; i < len(players); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			routine(collection, tracker, players[i])
		}(i)
	}

	wg.Wait()
	tracker.ListScores()
}

type GameOptions struct {
	PlayerCount int
	LogRounds   bool
	Rounds      int64
}

func declareWinner(tracker *scoretracker.ScoreTracker, playerMap map[string]*Player) {
	highestScore := 0
	scoreMapByScore := make(map[int][]string)

	fmt.Println(strings.Repeat("-", 80))

	for playerName, score := range tracker.PlayerScores {
		player := playerMap[playerName]
		fmt.Printf("Player %s (%d y.o.) from %s scored %d points!\n", player.Name, player.Age, player.Location, score)
		scoreMapByScore[score] = append(scoreMapByScore[score], player.Name)
		if score > highestScore {
			highestScore = score
		}
	}

	winners := scoreMapByScore[highestScore]

	if len(winners) > 1 {
		fmt.Printf("Tie! The winners are %s\n", winners)
	} else {
		fmt.Printf("The winner is %s!\n", winners[0])
	}
}

// Run blocks calling goroutine indefinitely until receives Interrupt or SIGTERM
func Run(gameOptions *GameOptions, controller controller.Controller) {
	controller.Announce("%d players!", gameOptions.PlayerCount)
	log.Default.Verbose = gameOptions.LogRounds

	players := []*Player{}
	playerMap := make(map[string]*Player)
	tracker := scoretracker.New()

	for i := 0; i < gameOptions.PlayerCount; i++ {
		controller.Announce("Player %d:", i+1)

		player := &Player{}

		controller.AskString("What's your name?", &player.Name)
		controller.AskString("Location?", &player.Location)
		controller.AskInt("How old are ya?", &player.Age)

		players = append(players, player)
		playerMap[player.Name] = player
		tracker.AddPlayer(player.Name)
	}

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		declareWinner(tracker, playerMap)
		os.Exit(1)
	}()

	controller.Announce("Running the game...Use ctrl+c to quit!")
	var i int64
	for i < gameOptions.Rounds {
		i++
		runOneRound(players, tracker, i)
	}
	declareWinner(tracker, playerMap)
}
