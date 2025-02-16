package view

import (
	"fishreports/controller"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ✅ Setup API routes
func SetupRoutes(router *gin.Engine, fishController *controller.FishSurveyController, countyController *controller.CountyController) {

	
	router.GET("/surveys", func(c *gin.Context) {
		// Expect species and county IDs instead of names.
		species := c.QueryArray("species") // species IDs
		minYear := c.Query("minYear")
		maxYear := c.Query("maxYear") // max year
		counties := c.QueryArray("counties") // county IDs
		lakes := c.QueryArray("lake") // lakes remain unchanged
		sortBy := c.Query("sort_by")
		order := c.Query("order")
		search := c.Query("search")
		limitStr := c.DefaultQuery("limit", "50")
		pageStr := c.DefaultQuery("page", "1")
		// New query parameter for game fish
		gameFishStr := c.DefaultQuery("game_fish", "false")
		gameFishOnly, _ := strconv.ParseBool(gameFishStr)

		limit, _ := strconv.Atoi(limitStr)
		page, _ := strconv.Atoi(pageStr)

		// Pass the parameters to the controller.
		filteredData := fishController.FilterAndSortData(
			species, minYear, maxYear, counties, lakes,
			sortBy, order, gameFishOnly, search, limit, page,
		)

		c.JSON(http.StatusOK, gin.H{
			"data":      filteredData["data"],
			"limit":     limit,
			"page":      page,
			"prev_page": filteredData["prev_page"],
			"next_page": filteredData["next_page"],
			"total":     filteredData["total"],
		})
	})


	router.GET("/graph", func(c *gin.Context) {
		dowNumber := c.Query("dow")
		speciesName := c.Query("species")
		surveyDate := c.Query("date")

		if dowNumber == "" || speciesName == "" || surveyDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing request query parameters: dow, species, or date"})
			return
		}

		graphData := fishController.GetFishCountData(dowNumber, speciesName, surveyDate)
		if graphData == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "No data found for the specified parameters"})
			return
		}

		c.JSON(http.StatusOK, graphData)
	})

	router.GET("/counties", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": countyController.GetCounties(),
		})
	})

	 // Route to get all species.
    router.GET("/species", func(c *gin.Context) {
        speciesList := fishController.GetAllSpecies()
        c.JSON(http.StatusOK, gin.H{
            "data": speciesList,
        })
    })

    // New endpoint to retrieve stats for a specific species by its ID.
    router.GET("/species/id/:species_id", func(c *gin.Context) {
    speciesID := c.Param("species_id")
    stats := fishController.GetSpeciesStatsByID(speciesID)
    if stats == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Species not found or no data available"})
        return
    }
    c.JSON(http.StatusOK, stats)
    })

    // New endpoint: GET /counties/id/:id
	router.GET("/counties/id/:id", func(c *gin.Context) {
		id := c.Param("id")
		county := countyController.GetCountyByID(id)
		if county == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "County not found"})
			return
		}
		stats := countyController.GetCountyStats(county)
		c.JSON(http.StatusOK, stats)
	})
}
