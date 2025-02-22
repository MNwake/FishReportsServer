package controller

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"math"
)

func (c *FishSurveyController) GetAllSpecies() []map[string]string {
    var speciesList []map[string]string

    // Iterate through the species map.
    for abbr, species := range c.Model.SpeciesMap {
        // Only include species if survey data exists for it.
        if !c.HasSurveyDataForSpecies(abbr) {
            continue
        }
        speciesList = append(speciesList, map[string]string{
            "id":             species.ID,
            "common_name":    species.CommonName,
            "image_url":      species.ImageURL,
            "description":    species.Description,
            "game_fish":      strconv.FormatBool(species.GameFish),
            "ScientificName": species.ScientificName,
            "SpeciesGroup":   species.SpeciesGroup,
        })
    }

    // Sort by common name.
    sort.Slice(speciesList, func(i, j int) bool {
        return speciesList[i]["common_name"] < speciesList[j]["common_name"]
    })

    return speciesList
}


// GetSpeciesStats aggregates statistics for a given species (by common name)
// across all surveys and returns county stats with integer percentages.
func (c *FishSurveyController) GetSpeciesStats(commonName string) map[string]interface{} {
	speciesAbbr := c.NormalizeSpecies(commonName)
	if speciesAbbr == "" {
		return nil
	}

	var (
		totalLengthSum int
		totalQuantity  int
		biggestLength  int
		shortestLength int = 1<<31 - 1 // max int value
	)

	// Global sets for lakes (for overall stats).
	lakesWithSpecies := make(map[string]bool)
	allLakes := make(map[string]bool)

	// Map to aggregate quantities by fish length for graph data.
	graphMap := make(map[int]int)

	// Maps to track lakes per county.
	allLakesByCounty := make(map[string]map[string]bool)
	speciesLakesByCounty := make(map[string]map[string]bool)

	// Iterate over all fish data by county.
	for _, fishDataList := range c.Model.FishDataByCounty {
		for _, data := range fishDataList {
			// Normalize the county name from the fish data.
			normalizedFishCounty := NormalizeCountyName(data.Result.CountyName)
			lakeName := data.Result.LakeName

			// Record this lake in the overall set.
			allLakes[strings.ToLower(lakeName)] = true

			// Use normalized county name as key for aggregation.
			if allLakesByCounty[normalizedFishCounty] == nil {
				allLakesByCounty[normalizedFishCounty] = make(map[string]bool)
			}
			allLakesByCounty[normalizedFishCounty][lakeName] = true

			// Check if this survey contains data for the species.
			for _, survey := range data.Result.Surveys {
				lengthData, exists := survey.Lengths[speciesAbbr]
				if !exists {
					continue
				}

				// Mark the lake as having the species.
				lakesWithSpecies[strings.ToLower(lakeName)] = true

				if speciesLakesByCounty[normalizedFishCounty] == nil {
					speciesLakesByCounty[normalizedFishCounty] = make(map[string]bool)
				}
				speciesLakesByCounty[normalizedFishCounty][lakeName] = true

				// Process each fish count entry.
				for _, count := range lengthData.FishCount {
					graphMap[count.Length] += count.Quantity
					totalLengthSum += count.Length * count.Quantity
					totalQuantity += count.Quantity

					if count.Length > biggestLength {
						biggestLength = count.Length
					}
					if count.Length < shortestLength {
						shortestLength = count.Length
					}
				}
			}
		}
	}

	// Debug: Log all aggregated fish data county keys.
	for key := range allLakesByCounty {
		log.Printf("Aggregated fish data county key: '%s'", key)
	}

	// Calculate weighted average length.
	avgLength := 0.0
	if totalQuantity > 0 {
		avgLength = float64(totalLengthSum) / float64(totalQuantity)
	} else {
		shortestLength = 0
	}

	// Calculate overall percentage of lakes with the species as an int.
	overallPercent := 0
	if len(allLakes) > 0 {
		overallPercent = int(math.Round((float64(len(lakesWithSpecies)) / float64(len(allLakes))) * 100))
	}

	// Convert aggregated graphMap to a slice of maps.
	aggregatedGraphData := []map[string]int{}
	for length, quantity := range graphMap {
		aggregatedGraphData = append(aggregatedGraphData, map[string]int{
			"length":   length,
			"quantity": quantity,
		})
	}

	sort.Slice(aggregatedGraphData, func(i, j int) bool {
		return aggregatedGraphData[i]["length"] < aggregatedGraphData[j]["length"]
	})

	// Build county stats: for each normalized county, compute the percentage of lakes with the species and include the county ID.
	var countyStats []map[string]interface{}
	for normalizedCounty, allLakesSet := range allLakesByCounty {
		totalLakes := len(allLakesSet)
		percentage := 0
		if totalLakes > 0 {
			speciesLakesCount := 0
			if speciesLakesByCounty[normalizedCounty] != nil {
				speciesLakesCount = len(speciesLakesByCounty[normalizedCounty])
			}
			percentage = int(math.Round((float64(speciesLakesCount) / float64(totalLakes)) * 100))
		}
		countyStats = append(countyStats, map[string]interface{}{
			"id":         getCountyID(normalizedCounty), // pass normalized key
			"percentage": percentage,
		})
	}

	sort.Slice(countyStats, func(i, j int) bool {
		return countyStats[i]["id"].(string) < countyStats[j]["id"].(string)
	})

	return map[string]interface{}{
		"species":         commonName,
		"percent_lakes":   overallPercent,
		"average_length":  avgLength,
		"biggest_length":  biggestLength,
		"shortest_length": shortestLength,
		"graph_data":      aggregatedGraphData,
		"total_fish":      totalQuantity,
		"counties":        countyStats,
	}
}

// GetSpeciesStatsByID finds the species by its ID and returns the aggregated stats.
func (c *FishSurveyController) GetSpeciesStatsByID(speciesID string) map[string]interface{} {
    var speciesKey string
    // Iterate over the species map (which is keyed by species code)
    for key, sp := range c.Model.SpeciesMap {
        if sp.ID == speciesID {
            speciesKey = key
            break
        }
    }
    if speciesKey == "" {
        return nil
    }
    // Retrieve the species using the found key.
    species := c.Model.SpeciesMap[speciesKey]
    // Now call the existing GetSpeciesStats using the species common name.
    return c.GetSpeciesStats(species.CommonName)
}

// HasSurveyDataForSpecies checks if any survey contains data for the given species abbreviation.
func (c *FishSurveyController) HasSurveyDataForSpecies(speciesAbbr string) bool {
    for _, fishDataList := range c.Model.FishDataByCounty {
        for _, data := range fishDataList {
            // Iterate through each survey in the county.
            for _, survey := range data.Result.Surveys {
                // If the species is found in the survey, return true.
                if _, exists := survey.Lengths[speciesAbbr]; exists {
                    return true
                }
            }
        }
    }
    return false
}