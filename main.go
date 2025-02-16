package main

import (
	"fishreports/model"
	"fishreports/controller"
	"fishreports/view"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the model.
	m := &model.FishSurveyModel{}

	// Declare a local variable for counties.
	var counties []model.County

	counties, err := controller.LoadCounties("data/minnesota_counties.json")
	if err != nil {
		log.Fatalf("Error loading counties: %v", err)
	}
	controller.Counties = counties 

	for _, county := range counties {
		log.Printf("Loaded county normalized: '%s' (original: '%s', ID: %s)", controller.NormalizeCountyName(county.CountyName), county.CountyName, county.ID)
	}

	// Load fish survey data.
	err = controller.LoadFishData(m, "data/surveys")
	if err != nil {
		log.Fatalf("Error loading fish survey data: %v", err)
	}

	// Load species metadata.
	err = controller.LoadSpeciesMap(m, "data/fish_species.json")
	if err != nil {
		log.Fatalf("Error loading species data: %v", err)
	}

	// Enhance counties with lake names from fish survey data.
	enhancedCounties := controller.EnhanceCountiesWithLakes(m, counties)

	// Create controllers.
	fishController := controller.NewFishSurveyController(m)
	countyController := controller.NewCountyController(enhancedCounties, m)

	// Setup router.
	router := gin.Default()
	view.SetupRoutes(router, fishController, countyController)

	log.Println("Server running on port 8080...")
	router.Run(":8080")
}