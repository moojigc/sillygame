package scoretracker

import (
	"fmt"
	"sync"

	"github.com/moojigc/routines_game/log"
)

type ScoreTracker struct {
	Mutex        sync.Mutex
	PlayerScores map[string]int
}

func (st *ScoreTracker) AddPlayer(player string) {
	st.PlayerScores[player] = 0
}

func (st *ScoreTracker) GetScoreByPlayer(player string) int {
	playerScore, ok := st.PlayerScores[player]
	if !ok {
		panic(fmt.Sprintf("No player named %s", player))
	}
	return playerScore
}

func (st *ScoreTracker) Increment(player string) {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	st.PlayerScores[player]++
}

func (st *ScoreTracker) Decrement(player string) {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	st.PlayerScores[player]--
}

func (st *ScoreTracker) ListScores() {
	st.Mutex.Lock()
	defer st.Mutex.Unlock()

	for player, score := range st.PlayerScores {
		log.Default.Print("Player: %s; Score: %d\t", player, score)
	}
	log.Default.Print("\n\n")
}

func New() *ScoreTracker {
	return &ScoreTracker{
		Mutex:        sync.Mutex{},
		PlayerScores: make(map[string]int),
	}
}
