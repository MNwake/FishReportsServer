package main

import (
	"fishreports/controller"
	"fishreports/model"
	"fishreports/view"

	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// ✅ Initialize the Model
	model := &model.FishSurveyModel{}

	// ✅ Load County Data
	err := controller.LoadCounties("data/minnesota_counties.json")
	if err != nil {
		log.Fatalf("Error loading county data: %v", err)
	}

	// ✅ Load Fish Survey Data
	err = controller.LoadFishData(model, "data/surveys")
	if err != nil {
		log.Fatalf("Error loading fish survey data: %v", err)
	}

	// ✅ Load Species Metadata
	err = controller.LoadSpeciesMap(model, "data/fish_species.json")
	if err != nil {
		log.Fatalf("Error loading species data: %v", err)
	}

	// ✅ Initialize Controller
	fishController := controller.NewFishSurveyController(model)

	// ✅ Setup Router
	router := gin.Default()
	view.SetupRoutes(router, fishController)

	log.Println("Server running on port 8080...")
	router.Run(":8080")
}
