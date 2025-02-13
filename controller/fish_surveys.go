package controller

import (
	"fishreports/model"

	"sort"
	"strconv"
	"strings"
)

// FishSurveyController provides methods for filtering, sorting,
// and paginating fish survey data.
type FishSurveyController struct {
	Model *model.FishSurveyModel
}

// NewFishSurveyController creates a new instance of FishSurveyController.
func NewFishSurveyController(model *model.FishSurveyModel) *FishSurveyController {
	return &FishSurveyController{Model: model}
}

// NormalizeSpecies converts a common species name to its abbreviation.
func (c *FishSurveyController) NormalizeSpecies(commonName string) string {
	commonName = strings.ToLower(commonName)
	for code, species := range c.Model.SpeciesMap {
		if strings.ToLower(species.CommonName) == commonName {
			return code
		}
	}
	return ""
}

// FilterAndSortData is the entry point for filtering, sorting, and paginating fish survey data.
func (c *FishSurveyController) FilterAndSortData(
	species, minYear, maxYear string,
	counties []string,
	sortBy, order string,
	gameFishOnly bool,
	search string,
	limit, page int,
) map[string]interface{} {
	var result []map[string]interface{}

	countySet := buildCountySet(counties)
	minYearInt := parseMinYear(minYear)
	maxYearInt := 0
	if maxYear != "" {
		maxYearInt, _ = strconv.Atoi(maxYear)
	}
	speciesAbbr := ""
	if species != "" {
		speciesAbbr = c.NormalizeSpecies(species)
	}

	// Iterate through fish data.
	for _, fishDataList := range c.Model.FishDataByCounty {
		for _, data := range fishDataList {
			if len(data.Result.Surveys) == 0 {
				continue
			}
			// Skip data for counties not in the filter.
			if len(counties) > 0 && !countySet[strings.ToLower(data.Result.CountyName)] {
				continue
			}
			for _, survey := range data.Result.Surveys {
				// Process each survey with the new parameters.
				rows := c.processSurvey(data, survey, speciesAbbr, minYearInt, maxYearInt, gameFishOnly, search)
				result = append(result, rows...)
			}
		}
	}

	// Set default sort options if not provided.
	if sortBy == "" {
		sortBy = "survey_date"
		order = "desc"
	}
	sortRows(result, sortBy, order)
	paginatedData, prevPage, nextPage := paginate(result, limit, page)

	return map[string]interface{}{
		"data":      paginatedData,
		"limit":     limit,
		"page":      page,
		"prev_page": prevPage,
		"next_page": nextPage,
		"total":     len(result),
	}
}

// buildCountySet converts a slice of county names to a lower-case map for quick lookup.
func buildCountySet(counties []string) map[string]bool {
	set := make(map[string]bool)
	for _, county := range counties {
		set[strings.ToLower(county)] = true
	}
	return set
}

// parseMinYear converts the minYear string into an integer.
func parseMinYear(minYear string) int {
	if minYear == "" {
		return 0
	}
	y, _ := strconv.Atoi(minYear)
	return y
}

func (c *FishSurveyController) processSurvey(
	data model.FishData,
	survey model.Survey,
	speciesAbbr string,
	minYearInt, maxYearInt int,
	gameFishOnly bool,
	search string,
) []map[string]interface{} {
	var rows []map[string]interface{}
	surveyYear := 0
	if len(survey.SurveyDate) >= 4 {
		surveyYear, _ = strconv.Atoi(survey.SurveyDate[:4])
	}
	// Only include surveys with surveyYear >= minYearInt.
	if minYearInt > 0 && surveyYear < minYearInt {
		return rows
	}
	// And if maxYearInt is provided, only include surveys with surveyYear <= maxYearInt.
	if maxYearInt > 0 && surveyYear > maxYearInt {
		return rows
	}

	for abbreviation, lengthData := range survey.Lengths {
		// Ensure species is set.
		if lengthData.Species == nil {
			if speciesObj, exists := c.Model.SpeciesMap[abbreviation]; exists {
				lengthData.Species = &speciesObj
			} else {
				continue
			}
		}
		// If gameFishOnly is true, skip non-game fish.
		if gameFishOnly && !lengthData.Species.GameFish {
			continue
		}
		// Skip if species filter is applied and doesn't match.
		if speciesAbbr != "" && abbreviation != speciesAbbr {
			continue
		}

		imageURL := ""
		if lengthData.Species != nil {
			imageURL = lengthData.Species.ImageURL
		}

		// Build the row.
		row := map[string]interface{}{
			"surveyID":        survey.SurveyID,
			"dow_number":      data.Result.DOWNumber,
			"survey_type":     survey.SurveyType,
			"survey_sub_type": survey.SurveySubType,
			"county_name":     data.Result.CountyName,
			"lake_name":       data.Result.LakeName,
			"survey_date":     survey.SurveyDate,
			"species_name":    lengthData.Species.CommonName,
			"image_url":       imageURL,
			"narrative":       survey.Narrative,
			"min_length":      lengthData.MinimumLength,
			"max_length":      lengthData.MaximumLength,
			"total_catch":     0,
		}

		// Calculate total catch.
		totalCatch := 0
		for _, summary := range survey.FishCatchSummaries {
			if summary.Species != nil && *summary.Species == abbreviation {
				if summary.TotalCatch != nil {
					totalCatch += *summary.TotalCatch
				}
			}
		}
		row["total_catch"] = totalCatch

		// Apply search filter if provided.
		if search != "" {
			lowerSearch := strings.ToLower(search)
			if !(strings.Contains(strings.ToLower(row["species_name"].(string)), lowerSearch) ||
				strings.Contains(strings.ToLower(row["county_name"].(string)), lowerSearch) ||
				strings.Contains(strings.ToLower(row["lake_name"].(string)), lowerSearch)) {
				continue
			}
		}

		rows = append(rows, row)
	}
	return rows
}

// sortRows sorts the rows based on the provided field and order.
func sortRows(rows []map[string]interface{}, sortBy, order string) {
	sort.Slice(rows, func(i, j int) bool {
		v1, ok1 := rows[i][sortBy]
		v2, ok2 := rows[j][sortBy]
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
			return false
		}
	})
}

// paginate returns the slice of rows for the requested page along with previous and next page numbers.
func paginate(rows []map[string]interface{}, limit, page int) ([]map[string]interface{}, int, int) {
	startIndex := (page - 1) * limit
	if startIndex >= len(rows) {
		return []map[string]interface{}{}, max(page-1, 1), page
	}
	endIndex := startIndex + limit
	if endIndex > len(rows) {
		endIndex = len(rows)
	}
	paginatedData := rows[startIndex:endIndex]
	return paginatedData, max(page-1, 1), page + 1
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}