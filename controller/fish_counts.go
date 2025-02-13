package controller

import (
	"log"
	"strconv"
)

// GetFishCountData retrieves fish count data based on the provided DOW, species name, and survey date.
func (c *FishSurveyController) GetFishCountData(dowStr, speciesName, surveyDate string) map[string]interface{} {
	log.Printf("🔍 GetFishCountData called with dowStr=%s, species=%s, surveyDate=%s", dowStr, speciesName, surveyDate)

	// Convert DOW to integer
	dow, err := strconv.Atoi(dowStr)
	if err != nil {
		log.Printf("❌ Invalid DOW number: %s", dowStr)
		return nil
	}

	// Normalize species name to abbreviation
	speciesAbbr := c.NormalizeSpecies(speciesName)
	log.Printf("✅ Normalized species '%s' to abbreviation '%s'", speciesName, speciesAbbr)
	if speciesAbbr == "" {
		log.Printf("❌ Species not found in SpeciesMap")
		return nil
	}

	// Iterate through fish data
	for _, fishDataList := range c.Model.FishDataByCounty {
		for _, data := range fishDataList {
			if data.Result.DOWNumber != dow {
				continue
			}


			for _, survey := range data.Result.Surveys {
				if survey.SurveyDate != surveyDate {
					continue
				}


				// Retrieve length data for the requested species
				lengthData, exists := survey.Lengths[speciesAbbr]
				if !exists {
					log.Printf("❌ No length data found for species: %s", speciesName)
					return nil
				}

				// Convert fish count to a structured response.
				fishCounts := make([]map[string]int, len(lengthData.FishCount))
				for i, count := range lengthData.FishCount {
					fishCounts[i] = map[string]int{
						"length":   count.Length,
						"quantity": count.Quantity,
					}
				}


				return map[string]interface{}{
					"species":    speciesName,
					"surveyDate": surveyDate,
					"data":       fishCounts,
				}
			}
		}
	}

	log.Printf("❌ No data found for given filters")
	return nil
}