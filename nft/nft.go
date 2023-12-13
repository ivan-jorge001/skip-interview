package nft

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Token struct {
	id    int
	attrs map[string]string
}

type RarityScorecard struct {
	Rarity float64
	Id     int
}

type Collection struct {
	Count int
	Url   string
}

type NFTAttributes struct {
	name  string
	value string
}

type Traits map[string]map[string]int

const URL = "https://go-challenge.skip.money"
const COLLECTION = "azuki"

func GetAllTraits(tokens []*Token) Traits {
	traits := Traits{}
	for _, v := range tokens {
		for key, value := range v.attrs {
			if traits[key] == nil {
				traits[key] = map[string]int{
					value: 1,
				}
			} else if _, ok := traits[key][value]; !ok {
				traits[key][value] = 1
			} else {
				traits[key][value] = traits[key][value] + 1
			}
		}
	}

	return traits
}

func CalculateRarity(tokens []*Token, traitsInfo map[string]map[string]int) []RarityScorecard {
	rarityScores := make([]RarityScorecard, len(tokens))

	for i, token := range tokens {
		if token == nil { // Check if the token is nil
			continue
		}

		var rarity float64
		for trait, value := range token.attrs {
			countWithTraitValue, exists := traitsInfo[trait][value]
			if exists && countWithTraitValue > 0 { // Ensure countWithTraitValue is not zero
				numValues := len(traitsInfo[trait])
				rarity += 1 / (float64(countWithTraitValue) * float64(numValues))
			}
		}
		rarityScores[i] = RarityScorecard{Rarity: rarity, Id: token.id}
	}

	return rarityScores
}

func FetchToken(tid int, colUrl string) (*Token, error) {
	url := fmt.Sprintf("%s/%s/%d.json", URL, colUrl, tid)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return &Token{}, fmt.Errorf("Error getting token %d :", tid)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &Token{}, fmt.Errorf("Error reading response for token %d :", tid)
	}

	attrs := make(map[string]string)
	if err := json.Unmarshal(body, &attrs); err != nil {
		return &Token{}, err
	}

	return &Token{
		id:    tid,
		attrs: attrs,
	}, nil
}
