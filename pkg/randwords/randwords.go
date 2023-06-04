package randwords

import (
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
)

type RandomWords struct {
	Adjectives []string `json:"adjectives"`
	Nouns      []string `json:"nouns"`
}

var randomWords RandomWords

// RandomPhrase returns a random adjective+noun phrase.
func RandomPhrase() string {
	randomWords, err := getRandomWords()
	if err != nil {
		return "RandomWords Error"
	}
	randAdj := randomWords.Adjectives[random(len(randomWords.Adjectives))]
	randNoun := randomWords.Nouns[random(len(randomWords.Nouns))]

	return capitalize(randAdj) + capitalize(randNoun)
}

func capitalize(word string) string {
	if len(word) == 0 {
		return ""
	}
	firstLetter := strings.ToUpper(string(word[0]))
	return firstLetter + word[1:]
}

// getRandomWords loads the random words from the json files or returns the cached version.
func getRandomWords() (*RandomWords, error) {
	if len(randomWords.Adjectives) != 0 {
		return &randomWords, nil
	}

	randomWordsJsonPath := os.Getenv("RANDOM_WORDS_PATH")
	if randomWordsJsonPath == "" {
		randomWordsJsonPath = filepath.Join("pkg", "randwords", "static", "random_words.json")
	}

	bytes, err := ioutil.ReadFile(randomWordsJsonPath)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, &randomWords); err != nil {
		return nil, err
	}

	return &randomWords, nil
}

// random returns a random number between 0 and max.
func random(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return int(n.Int64())
}

func init() { getRandomWords() }
