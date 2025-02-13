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

	// Load counties.
	counties, err := controller.LoadCounties("data/minnesota_counties.json")
	if err != nil {
		log.Fatalf("Error loading county data: %v", err)
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
	countyController := controller.NewCountyController(enhancedCounties)

	// Setup router.
	router := gin.Default()
	// Pass both controllers to your view setup (update your SetupRoutes accordingly).
	view.SetupRoutes(router, fishController, countyController)

	log.Println("Server running on port 8080...")
	router.Run(":8080")
}

