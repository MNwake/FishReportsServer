package view

import (
	"fishreports/controller"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ✅ Setup API routes
func SetupRoutes(router *gin.Engine, fishController *controller.FishSurveyController, countyController *controller.CountyController) {

	router.GET("/data", func(c *gin.Context) {
    species := c.Query("species")
    minYear := c.Query("minYear")
    maxYear := c.Query("maxYear") // New query parameter for max year
    counties := c.QueryArray("county")
    sortBy := c.Query("sort_by")
    order := c.Query("order")
    search := c.Query("search") // New search parameter
    limitStr := c.DefaultQuery("limit", "50")
    pageStr := c.DefaultQuery("page", "1")
    
    // New query parameter for game fish
    gameFishStr := c.DefaultQuery("game_fish", "false")
    gameFishOnly, _ := strconv.ParseBool(gameFishStr)

    limit, _ := strconv.Atoi(limitStr)
    page, _ := strconv.Atoi(pageStr)

    // Pass the new parameters to the controller.
    filteredData := fishController.FilterAndSortData(species, minYear, maxYear, counties, sortBy, order, gameFishOnly, search, limit, page)

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

	// Add this new route to SetupRoutes in api_routes.go
	router.GET("/species", func(c *gin.Context) {
		speciesList := fishController.GetAllSpecies()
		c.JSON(http.StatusOK, gin.H{
			"data": speciesList,
	})
})
}
