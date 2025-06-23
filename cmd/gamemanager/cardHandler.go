package gamemanager

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type Alias struct {
  Set string  `json:"set"`
  ID  uint    `json:"id"`
}

type StaticCardDataRaw struct {
  Name          string      `json:"name,omitempty"`
  ImageSrc      string      `json:"imageSrc"`
  Alias         Alias       `json:"alias,omitempty"`
  PreCondition  *Expression  `json:"preCondition,omitempty"`
  Effect        *CardEffect `json:"effect,omitempty"`
  CardType      string      `json:"cardType,omitempty"`
}

func (cd *StaticCardDataRaw) UnmarshalJSON(data []byte) error {
  cd.Alias = Alias{
    Set: "",
    ID: 0,
  } // set default value before unmarshaling

  // create alias to prevent endless loop
  type TMP StaticCardDataRaw
  tmp := (*TMP)(cd)

  return json.Unmarshal(data, tmp)
}

type StaticCardData struct {
  Name          string
  ImageSrc      string
  Alias         *StaticCardData
  PreCondition  *Expression
  Effect        *CardEffect
  CardType      string
}

type CardHandler struct {
  cardLookup map[string]([]StaticCardData)
}

// Takes a string representing the available cards and returns a 
// cardHandler with set1 set to those cards
func SetupFromString(content string) *CardHandler {
	cardHandler := &CardHandler{
		cardLookup: make(map[string][]StaticCardData, 1),
	}
	rawLookupTables := make(map[string][]StaticCardDataRaw)

	if err := processSet("set1", []byte(content), cardHandler, rawLookupTables); err != nil {
		fmt.Printf("Error processing string content: %v\n", err)
	}

	resolveAliases(cardHandler, rawLookupTables)
	return cardHandler
}

// Takes a directory path and returns a cardHandler generated from the
// text files contained in it
func SetupFromDirectory(path string) *CardHandler {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	cardHandler := &CardHandler{
		cardLookup: make(map[string][]StaticCardData, len(entries)),
	}
	rawLookupTables := make(map[string][]StaticCardDataRaw)

	for _, e := range entries {
		fileName := e.Name()
		text, err := os.ReadFile(path + "/" + fileName)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", fileName, err)
			continue
		}

		setName := strings.Split(fileName, ".")[0]
		if err := processSet(setName, text, cardHandler, rawLookupTables); err != nil {
			fmt.Printf("Error processing set from file %s: %v\n", fileName, err)
		}
	}

	resolveAliases(cardHandler, rawLookupTables)
	return cardHandler
}

// processSet handles unmarshalling and initial processing of a single set.
func processSet(setName string, content []byte, ch *CardHandler, rawLookups map[string][]StaticCardDataRaw) error {
	var setLookupTableRaw []StaticCardDataRaw
	err := json.Unmarshal(content, &setLookupTableRaw)
	if err != nil {
		return fmt.Errorf("failed to unmarshal set %s: %w", setName, err)
	}

	rawLookups[setName] = setLookupTableRaw
	ch.cardLookup[setName] = make([]StaticCardData, 0, len(setLookupTableRaw))

	for _, element := range setLookupTableRaw {
		ch.cardLookup[setName] = append(ch.cardLookup[setName], StaticCardData{
			Name:         element.Name,
			ImageSrc:     element.ImageSrc,
			Alias:        nil,
			PreCondition: element.PreCondition,
			Effect:       element.Effect,
      CardType:     element.CardType,
		})
	}
	return nil
}

// resolveAliases populates the Alias fields after all cards have been loaded.
func resolveAliases(ch *CardHandler, rawLookups map[string][]StaticCardDataRaw) {
	for setName, rawLookupTable := range rawLookups {
		for index, element := range rawLookupTable {
			if element.Alias.Set == "" {
				continue
			}
			if _, ok := ch.cardLookup[element.Alias.Set]; !ok {
				fmt.Println("Loading in card with alias pointing to unknown set", element.Alias.Set)
				continue
			}
			if uint(len(ch.cardLookup[element.Alias.Set])) <= element.Alias.ID {
				fmt.Println("Loading in card with alias pointing to valid set but unknown id", element.Alias.ID)
				continue
			}

			ch.cardLookup[setName][index].Alias = &ch.cardLookup[element.Alias.Set][element.Alias.ID]
		}
	}
}
