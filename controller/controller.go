package controller

import (
	"fishreports/model"

	"log"
	"sort"
	"strconv"
	"strings"
)

// ✅ FishSurveyController struct
type FishSurveyController struct {
	Model *model.FishSurveyModel
}

// ✅ Initialize Controller
func NewFishSurveyController(model *model.FishSurveyModel) *FishSurveyController {
	return &FishSurveyController{Model: model}
}


// ✅ Normalize Species Name to Abbreviation
func (c *FishSurveyController) NormalizeSpecies(commonName string) string {
	commonName = strings.ToLower(commonName)
	for code, species := range c.Model.SpeciesMap {
		if strings.ToLower(species.CommonName) == commonName {
			return code
		}
	}
	return ""
}

// // ✅ Safe String-to-Integer Conversion
// func atoiSafe(str string) int {
// 	value, err := strconv.Atoi(str)
// 	if err != nil {
// 		return 0
// 	}
// 	return value
// }

// ✅ Filter & Sort Fish Data
func (c *FishSurveyController) FilterAndSortData(species string, minYear string, counties []string, sortBy string, order string, limit int, page int) map[string]interface{} {
	var result []map[string]interface{}

	// Convert counties to a map for quick lookup
	countySet := make(map[string]bool)
	for _, county := range counties {
		countySet[strings.ToLower(county)] = true
	}

	// Convert minYear to integer
	minYearInt := 0
	if minYear != "" {
		minYearInt, _ = strconv.Atoi(minYear)
	}

	// Normalize species name to abbreviation
	speciesAbbr := ""
	if species != "" {
		speciesAbbr = c.NormalizeSpecies(species)
	}

	// Iterate through all fish data by county
	for _, fishDataList := range c.Model.FishDataByCounty {
		for _, data := range fishDataList {
			if len(data.Result.Surveys) == 0 {
				continue
			}

			// Skip counties not in the filter
			if len(counties) > 0 && !countySet[strings.ToLower(data.Result.CountyName)] {
				continue
			}

			for _, survey := range data.Result.Surveys {
				surveyYear := 0
				if len(survey.SurveyDate) >= 4 {
					surveyYear, _ = strconv.Atoi(survey.SurveyDate[:4])
				}
				if minYearInt > 0 && surveyYear < minYearInt {
					continue
				}

				for abbreviation, lengthData := range survey.Lengths {
					// Ensure `Species` is not nil by fetching from speciesMap
					if lengthData.Species == nil {
						if speciesObj, exists := c.Model.SpeciesMap[abbreviation]; exists {
							lengthData.Species = &speciesObj
						} else {
							continue
						}
					}

					// Skip species not matching the filter
					if speciesAbbr != "" && abbreviation != speciesAbbr {
						continue
					}

					imageURL := ""
					if lengthData.Species != nil {
						imageURL = lengthData.Species.ImageURL
					}

					description := ""
					if lengthData.Species != nil {
						description = lengthData.Species.Description
					}

					row := map[string]interface{}{
						"surveyID":     survey.SurveyID,
						"dow_number":   data.Result.DOWNumber,
						"county_name":  data.Result.CountyName,
						"lake_name":    data.Result.LakeName,
						"survey_date":  survey.SurveyDate,
						"species_name": lengthData.Species.CommonName,
						"image_url":    imageURL,
						"description":  description,
						"min_length":   lengthData.MinimumLength,
						"max_length":   lengthData.MaximumLength,
						"total_catch":  0,
					}

					// Calculate total catch
					for _, summary := range survey.FishCatchSummaries {
						if summary.Species != nil && *summary.Species == abbreviation {
							if summary.TotalCatch != nil {
								row["total_catch"] = row["total_catch"].(int) + *summary.TotalCatch
							}
						}
					}

					result = append(result, row)
				}
			}
		}
	}

	// ✅ **Apply Default Sorting Order**
	if sortBy == "" {
		sortBy = "survey_date" // Default sort by survey date
		order = "desc"         // Default to descending order (most recent first)
	}

	// ✅ **Sort Data**
	sort.Slice(result, func(i, j int) bool {
		// First check if the fields exist and are of the expected type
		v1, ok1 := result[i][sortBy]
		v2, ok2 := result[j][sortBy]
		
		// If either value is missing, treat them as "less than"
		if !ok1 || !ok2 {
			return ok2
		}

		switch sortBy {
		case "total_catch":
			n1, ok1 := v1.(int)
			n2, ok2 := v2.(int)
			if !ok1 || !ok2 {
				return false
			}
			if order == "asc" {
				return n1 < n2
			}
			return n1 > n2
		case "survey_date", "lake_name", "county_name", "species_name":
			s1, ok1 := v1.(string)
			s2, ok2 := v2.(string)
			if !ok1 || !ok2 {
				return false
			}
			if order == "asc" {
				return s1 < s2
			}
			return s1 > s2
		case "max_length", "min_length":
			n1, ok1 := v1.(int)
			n2, ok2 := v2.(int)
			if !ok1 || !ok2 {
				return false
			}
			if order == "asc" {
				return n1 < n2
			}
			return n1 > n2
		default:
			// If sortBy is invalid, maintain stable order
			return false
		}
	})

	// ✅ **Apply Pagination**
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if startIndex >= len(result) {
		return map[string]interface{}{
			"data":      []map[string]interface{}{},
			"limit":     limit,
			"page":      page,
			"prev_page": max(page-1, 1),
			"next_page": page,
			"total":     len(result),
		}
	}
	if endIndex > len(result) {
		endIndex = len(result)
	}

	paginatedData := result[startIndex:endIndex]

	// ✅ **Return paginated response**
	return map[string]interface{}{
		"data":      paginatedData,
		"limit":     limit,
		"page":      page,
		"prev_page": max(page-1, 1),
		"next_page": page + 1,
		"total":     len(result),
	}
}

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

			log.Printf("✅ Found lake for DOW: %s (%s)", data.Result.LakeName, data.Result.CountyName)

			for _, survey := range data.Result.Surveys {
				if survey.SurveyDate != surveyDate {
					continue
				}

				log.Printf("✅ Found matching survey date: %s", survey.SurveyDate)

				// Retrieve length data for the requested species
				lengthData, exists := survey.Lengths[speciesAbbr]
				if !exists {
					log.Printf("❌ No length data found for species: %s", speciesName)
					return nil
				}

				// Convert fish count to structured response
				fishCounts := make([]map[string]int, len(lengthData.FishCount))
				for i, count := range lengthData.FishCount {
					fishCounts[i] = map[string]int{
						"length":   count.Length,
						"quantity": count.Quantity,
					}
				}

				log.Printf("📊 Returning fish count data: %v", fishCounts)

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

// func compareNumbers(a interface{}, b interface{}, order string) bool {
// 	aInt, _ := a.(int)
// 	bInt, _ := b.(int)
// 	if order == "desc" {
// 		return aInt > bInt
// 	}
// 	return aInt < bInt
// }

// func compareStrings(a string, b string, order string) bool {
// 	if order == "desc" {
// 		return a > b
// 	}
// 	return a < b
// }

// Add this new method to FishSurveyController
func (c *FishSurveyController) GetAllSpecies() []map[string]string {
	var speciesList []map[string]string

	// Convert map to sorted slice
	for _, species := range c.Model.SpeciesMap {
		speciesList = append(speciesList, map[string]string{
			"common_name": species.CommonName,
			"image_url":   species.ImageURL,
			"game_fish":   strconv.FormatBool(species.GameFish),
		})
	}

	// Sort by common name
	sort.Slice(speciesList, func(i, j int) bool {
		return speciesList[i]["common_name"] < speciesList[j]["common_name"]
	})

	return speciesList
}
