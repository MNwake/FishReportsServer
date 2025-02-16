package controller

import (
	"fishreports/model"
	"log"
	"sort"
	"strings"
	"math"

)



// CountyController handles county-related operations.
type CountyController struct {
	Counties []model.County
	FishSurveyModel  *model.FishSurveyModel 
}


// In controller/counties.go
func NewCountyController(counties []model.County, fishModel *model.FishSurveyModel) *CountyController {
    return &CountyController{
        Counties:         counties,
        FishSurveyModel:  fishModel,
    }
}

// GetCounties returns the stored counties.
func (cc *CountyController) GetCounties() []model.County {
	return cc.Counties
}

func NormalizeCountyName(name string) string {
    name = strings.ToLower(strings.TrimSpace(name))
    // Remove common punctuation.
    name = strings.ReplaceAll(name, ".", "")
    name = strings.ReplaceAll(name, ",", "")
    // Remove suffix " county" if present.
    if strings.HasSuffix(name, " county") {
        name = strings.TrimSuffix(name, " county")
        name = strings.TrimSpace(name)
    }
    return name
}

// EnhanceCountiesWithLakes enriches the given counties slice with lake names from the fish survey data.
// It returns the updated slice.
func EnhanceCountiesWithLakes(m *model.FishSurveyModel, counties []model.County) []model.County {
	// Create a mapping from normalized county name to a set of lake names.
	lakesByCounty := make(map[string]map[string]bool)
	for _, fishDataList := range m.FishDataByCounty {
		for _, data := range fishDataList {
			countyKey := NormalizeCountyName(data.Result.CountyName)
			lakeName := data.Result.LakeName
			if lakesByCounty[countyKey] == nil {
				lakesByCounty[countyKey] = make(map[string]bool)
			}
			lakesByCounty[countyKey][lakeName] = true
		}
	}

	// Enrich each county in the slice.
	for i, county := range counties {
		normalizedCounty := NormalizeCountyName(county.CountyName)
		if lakeSet, exists := lakesByCounty[normalizedCounty]; exists {
			var lakes []string
			for lake := range lakeSet {
				lakes = append(lakes, lake)
			}
			sort.Strings(lakes)
			counties[i].Lakes = lakes
		} else {
			log.Printf("County '%s' (normalized: '%s') has no associated lakes in fish data", county.CountyName, normalizedCounty)
		}
	}
	return counties
}

func getCountyID(countyName string) string {
    normalizedInput := NormalizeCountyName(countyName)
    // Handle known discrepancies:
    if normalizedInput == "st louis" {
        normalizedInput = "saint louis"
    }
    for _, county := range Counties {
        if NormalizeCountyName(county.CountyName) == normalizedInput {
            return county.ID
        }
    }
    return ""
}

// GetCountyByID searches for a county with the matching ID.
// Returns a pointer to the county if found, or nil otherwise.
func (cc *CountyController) GetCountyByID(id string) *model.County {
	for i := range cc.Counties {
		if cc.Counties[i].ID == id {
			return &cc.Counties[i]
		}
	}
	return nil
}

// GetCountyStats computes statistics for the given county.
// It aggregates data from the fish survey model using a normalized county name match.
func (cc *CountyController) GetCountyStats(county *model.County) map[string]interface{} {
	stats := make(map[string]interface{})

	// Base county info and lakes count.
	stats["county"] = county
	stats["number_of_lakes"] = len(county.Lakes)

	normalizedCounty := NormalizeCountyName(county.CountyName)
	var surveys []model.FishData

	// Check if FishSurveyModel or its FishDataByCounty is nil.
	if cc.FishSurveyModel == nil || cc.FishSurveyModel.FishDataByCounty == nil {
		// No fish data available; return base stats.
		stats["survey_ids"] = []string{}
		stats["total_surveys"] = 0
		stats["total_fish_caught"] = 0
		stats["number_of_species"] = 0
		stats["species_distribution"] = map[string]float64{}
		stats["average_fish_per_survey"] = 0.0
		return stats
	}

	// Aggregate fish survey data that match the normalized county name.
	for key, fishDataList := range cc.FishSurveyModel.FishDataByCounty {
		if NormalizeCountyName(key) == normalizedCounty {
			surveys = append(surveys, fishDataList...)
		}
	}

	var surveyIDs []string
	speciesCounts := make(map[string]int)
	totalFishCaught := 0
	totalSurveys := 0

	// Process each survey.
	for _, fishData := range surveys {
		for _, survey := range fishData.Result.Surveys {
			surveyIDs = append(surveyIDs, survey.SurveyID)
			totalSurveys++
			// Process each fish catch summary.
			for _, summary := range survey.FishCatchSummaries {
				if summary.Species != nil && summary.TotalCatch != nil {
					// Use species abbreviation from the summary.
					speciesAbbrev := *summary.Species
					speciesID := speciesAbbrev
					// Attempt to look up the species ID from the SpeciesMap.
					if speciesInfo, exists := cc.FishSurveyModel.SpeciesMap[speciesAbbrev]; exists {
						speciesID = speciesInfo.ID
					}
					count := *summary.TotalCatch
					speciesCounts[speciesID] += count
					totalFishCaught += count
				}
			}
		}
	}

	stats["survey_ids"] = surveyIDs
	stats["total_surveys"] = totalSurveys
	stats["total_fish_caught"] = totalFishCaught
	stats["number_of_species"] = len(speciesCounts)

	// Build pie chart data: percentage distribution per species (using species IDs).
	speciesDistribution := make(map[string]float64)
	if totalFishCaught > 0 {
		for speciesID, count := range speciesCounts {
			percentage := (float64(count) / float64(totalFishCaught)) * 100.0
			// Round to 2 decimal places.
			percentage = math.Round(percentage*100) / 100
			speciesDistribution[speciesID] = percentage
		}
	}
	stats["species_distribution"] = speciesDistribution

	// Additional stat: average fish caught per survey.
	if totalSurveys > 0 {
		avg := float64(totalFishCaught) / float64(totalSurveys)
		// Round average to 2 decimal places.
		avg = math.Round(avg*100) / 100
		stats["average_fish_per_survey"] = avg
	} else {
		stats["average_fish_per_survey"] = 0.0
	}

	return stats
}