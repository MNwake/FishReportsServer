package controller

import (
	"sort"
	"strconv"
)

// GetAllSpecies returns a sorted slice of species details.
func (c *FishSurveyController) GetAllSpecies() []map[string]string {
	var speciesList []map[string]string

	// Convert the species map to a slice.
	for _, species := range c.Model.SpeciesMap {
		speciesList = append(speciesList, map[string]string{
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