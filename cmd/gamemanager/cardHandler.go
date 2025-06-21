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
  PreCondition  Expression  `json:"preCondition,omitempty"`
  Effect        CardEffect  `json:"effect,omitempty"`
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
  PreCondition  Expression
  Effect        CardEffect
}

type CardHandler struct {
  cardLookup map[string]([]StaticCardData)
}

func SetupFromDirectory(path string) *CardHandler {
  entries, err := os.ReadDir(path)
  if err != nil {
    log.Fatal(err)
  }

  var cardHandler CardHandler = CardHandler{
    cardLookup: make(map[string][]StaticCardData, len(entries)),
  }
  
  rawLookupTables := make(map[string]([]StaticCardDataRaw))

  // populate raw lookup tables
  for _, e := range entries {
    setName := e.Name()

    text, err := os.ReadFile(path + "/" + setName)

    setName = strings.Split(setName, ".")[0]

    if err != nil {
      fmt.Println(err)
    }
    fmt.Print(string(text))

    var setLookupTableRaw []StaticCardDataRaw

    err = json.Unmarshal(text, &setLookupTableRaw)
    if err != nil {
      fmt.Println(err)
    }

    rawLookupTables[setName] = setLookupTableRaw

    cardHandler.cardLookup[setName] = make([]StaticCardData, 0, len(setLookupTableRaw))

    // generate actual lookup tables from raw lookup tables
    for _, element := range setLookupTableRaw {
      cardHandler.cardLookup[setName] = append(cardHandler.cardLookup[setName], StaticCardData{
        Name: element.Name,
        ImageSrc: element.ImageSrc,
        Alias: nil,
        PreCondition: element.PreCondition,
        Effect: element.Effect,
      })
    }
  }

  // fill out aliases  
  for setName, rawLookupTable := range rawLookupTables {
    for index, element := range rawLookupTable {
      if element.Alias.Set != "" {
        cardHandler.cardLookup[setName][index].Alias = &cardHandler.cardLookup[element.Alias.Set][element.Alias.ID]
      }
    }
  }

  return &cardHandler
}
