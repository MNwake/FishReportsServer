package model

import (
	"sync"
)

// ✅ Data structures
type Species struct {
	Code           string `json:"code"`
	CommonName     string `json:"common_name"`
	ScientificName string `json:"scientific_name"`
	GameFish       bool   `json:"game_fish"`
	SpeciesGroup   string `json:"species_group"`
	ImageURL       string `json:"image_url"`
	Description    string `json:"description"`
}

// ✅ Struct for fishCount entry
type FishCount struct {
	Length   int `json:"length"`
	Quantity int `json:"quantity"`
}

// ✅ Struct for fish length data
type LengthData struct {
	Species       *Species    `json:"species"`
	MinimumLength int         `json:"minimum_length"`
	MaximumLength int         `json:"maximum_length"`
	FishCount     []FishCount `json:"fishCount"` // 🔄 Fixed to accept an array of FishCount
}


type Survey struct {
	SurveyID           string                 `json:"surveyID"`
	SurveyDate         string                 `json:"surveyDate"`
	FishCatchSummaries []FishCatchSummary     `json:"fishCatchSummaries"`
	Narrative          string                 `json:"narrative"`  // <-- New narrative field
	Lengths            map[string]*LengthData `json:"lengths"`
	SurveyType 		   string				  `json:"surveyType"`
	SurveySubType	   string				  `json:"suveySubType"`
}

type FishCatchSummary struct {
	Species    *string `json:"species"`
	TotalCatch *int    `json:"totalCatch"`
}

type FishData struct {
	Result struct {
		DOWNumber  int      `json:"DOWNumber"`
		CountyName string   `json:"countyName"`
		LakeName   string   `json:"lakeName"`
		Surveys    []Survey `json:"surveys"`
	} `json:"result"`
}

// ✅ FishSurveyModel: Stores loaded fish data & species mapping
type FishSurveyModel struct {
	FishDataByCounty map[string][]FishData
	SpeciesMap       map[string]Species
	Mutex            sync.Mutex
}



// County struct represents the county data.
type County struct {
	CountyName  string   `json:"county_name"`
	FIPSCode    string   `json:"fips_code"`
	CountySeat  string   `json:"county_seat"`
	Established int      `json:"established"`
	Origin      string   `json:"origin"`
	Etymology   string   `json:"etymology"`
	Population  int      `json:"population"`
	AreaSqMiles float64  `json:"area_sq_miles"`
	MapImageURL string   `json:"map_image_url"`
	Lakes		[]string `json:"lakes,omitempty"`
}