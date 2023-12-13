package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/skip-money/coding-challenge/nft"
	"github.com/skip-money/coding-challenge/scheduler"
)

const COLLECTION = "azuki"
const COLOR_GREEN = "\033[32m"
const COLOR_RED = "\033[31m"
const COLOR_RESET = "\033[0m"

var logger *log.Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
var concurrentRequests int32

func WriteRarityScoresToFile(rarityScores []nft.RarityScorecard, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // For pretty-printing
	return encoder.Encode(rarityScores)
}

func main() {
	fmt.Println(time.Now())
	azuki := nft.Collection{
		Count: 10000,
		Url:   "azuki1",
	}

	jobs := []*scheduler.Job[*nft.Token]{}
	for i := 0; i < azuki.Count; i++ {

		fetch := func(id int, name string) func(ctx context.Context) (*nft.Token, error) {
			return func(ctx context.Context) (*nft.Token, error) {
				return nft.FetchToken(id, name)
			}
		}

		jobs = append(jobs, &scheduler.Job[*nft.Token]{
			Request: fetch(i, azuki.Url),
		})

	}
	sch := scheduler.CreateScheduler[*nft.Token](jobs)
	res := sch.RunExponentialBackOff()
	tokens := []*nft.Token{}
	for _, v := range res {
		tokens = append(tokens, *v.Response)
	}

	rarity := nft.CalculateRarity(tokens, nft.GetAllTraits(tokens))
	fmt.Println(time.Now())

	err := WriteRarityScoresToFile(rarity, "rarity_scores.json")
	if err != nil {
		fmt.Printf("Error writing JSON file: %s\n", err)
	} else {
		fmt.Println("Rarity scores written to 'rarity_scores.json'")
	}

	os.Exit(0)
}
