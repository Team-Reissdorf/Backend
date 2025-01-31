package performanceManagement

import "github.com/gin-gonic/gin"

// CreatePerformance creates a new performance entry
// @Summary Creates a new performance entry
// @Description Creates a new performance entry with the given data. Maximum 3 entries/discipline/athlete/day are allowed (performance limit).
// @Tags Performance Management
// @Accept json
// @Produce json
// @Param Performance body PerformanceBody true "Details of a performance"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 409 {object} endpoints.ErrorResponse "Performance limit reached"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/create [post]
func CreatePerformance(c *gin.Context) {

}
