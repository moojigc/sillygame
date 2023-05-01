package scoretracker

import (
	"fmt"
	"sync"

	"github.com/moojigc/routines_game/log"
)

type ScoreTracker struct {
	Mutex        sync.Mutex
	playerScores map[string]int
}

func (st *ScoreTracker) AddPlayer(player string) {
	st.playerScores[player] = 0
}

func (st *ScoreTracker) GetScoreByPlayer(player string) int {
	playerScore, ok := st.playerScores[player]
	if !ok {
		panic(fmt.Sprintf("No player named %s", player))
	}
	return playerScore
}

func (st *ScoreTracker) Increment(player string) {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	st.playerScores[player]++
}

func (st *ScoreTracker) Decrement(player string) {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	st.playerScores[player]--
}

func (st *ScoreTracker) ListScores() {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	for player, score := range st.playerScores {
		log.Default.Print("Player: %s; Score: %d\t", player, score)
	}
	log.Default.Print("\n\n")
}

func (st *ScoreTracker) DeclareWinner() {
	highestScore := 0
	scoreMapByScore := make(map[int][]string)

	fmt.Println("------------------------------------")

	for player, score := range st.playerScores {
		fmt.Printf("Player %s scored %d points!\n", player, score)
		scoreMapByScore[score] = append(scoreMapByScore[score], player)
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

func New() *ScoreTracker {
	return &ScoreTracker{
		Mutex:        sync.Mutex{},
		playerScores: make(map[string]int),
	}
}
