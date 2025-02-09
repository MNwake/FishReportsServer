package view

import (
	"fishreports/controller"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ✅ Setup API routes
func SetupRoutes(router *gin.Engine, fishController *controller.FishSurveyController) {

	router.GET("/data", func(c *gin.Context) {
		species := c.Query("species")
		minYear := c.Query("minYear")
		counties := c.QueryArray("county")
		sortBy := c.Query("sort_by")
		order := c.Query("order")
		limitStr := c.DefaultQuery("limit", "50")
		pageStr := c.DefaultQuery("page", "1")

		// Convert to integers
		limit, _ := strconv.Atoi(limitStr)
		page, _ := strconv.Atoi(pageStr)

		// ✅ Get paginated, sorted data
		filteredData := fishController.FilterAndSortData(species, minYear, counties, sortBy, order, limit, page)

		// ✅ Return response
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
			"data": controller.GetAllCounties(),
		})
	})
}
