package controller

import (
	"fishreports/model"
	"log"
	"sort"
	"strings"
)



// CountyController handles county-related operations.
type CountyController struct {
	Counties []model.County
}

// NewCountyController creates a new CountyController.
func NewCountyController(counties []model.County) *CountyController {
	return &CountyController{Counties: counties}
}

// GetCounties returns the stored counties.
func (cc *CountyController) GetCounties() []model.County {
	return cc.Counties
}

// normalizeCountyName returns a lower-case county name and removes any trailing " county".
func normalizeCountyName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
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
			countyKey := normalizeCountyName(data.Result.CountyName)
			lakeName := data.Result.LakeName
			if lakesByCounty[countyKey] == nil {
				lakesByCounty[countyKey] = make(map[string]bool)
			}
			lakesByCounty[countyKey][lakeName] = true
		}
	}

	// Enrich each county in the slice.
	for i, county := range counties {
		normalizedCounty := normalizeCountyName(county.CountyName)
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