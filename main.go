package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	col "github.com/moojigc/routines_game/collection"
	log "github.com/moojigc/routines_game/log"
	"github.com/moojigc/routines_game/scoretracker"
	st "github.com/moojigc/routines_game/scoretracker"
)

type Person struct {
	Name     string
	Location string
	Age      string
}

func randInt(x int64) int64 {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(x))
	return nBig.Int64()
}

func routine(c *col.Collection, tracker *st.ScoreTracker, p Person) {
	time.Sleep(time.Microsecond * time.Duration(randInt(10)))
	c.Add("name", p.Name).
		Add("location", p.Location).
		Add("age", p.Age)

	if !c.Has("face") {
		c.Add("face", "cute")
		tracker.Increment(p.Name)
	}
}

func runTest(people [2]Person, tracker *st.ScoreTracker, round int) {
	log.Default.Print("Round %d\n", round)

	var wg sync.WaitGroup

	collection := col.New()
	randSelector := randInt(2)

	if randSelector == 0 {
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				routine(collection, tracker, people[i])
			}(i)
		}
	} else {
		for i := 1; i >= 0; i-- {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				routine(collection, tracker, people[i])
			}(i)
		}
	}

	wg.Wait()
	tracker.ListScores()
}

func main() {
	flag.BoolVar(&log.Default.Include, "verbose", false, "Log each result?")
	flag.Parse()

	people := [2]Person{
		{
			Name:     "",
			Location: "CT",
			Age:      "27",
		},
		{
			Name:     "",
			Location: "CT",
			Age:      "25",
		},
	}

	fmt.Println("Player 1:")
	fmt.Scanln(&people[0].Name)
	fmt.Println("Player 2:")
	fmt.Scanln(&people[1].Name)

	tracker := scoretracker.New()
	tracker.AddPlayer(people[0].Name)
	tracker.AddPlayer(people[1].Name)

	// for i := 0; i < 50; i++ {
	// 	runTest(tracker, i+1)
	// }
	// tracker.DeclareWinner()

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		tracker.DeclareWinner()
		os.Exit(1)
	}()

	fmt.Println("Running the game...")
	i := 0
	for {
		i++
		runTest(people, tracker, i)
	}
}
