package ping

import (
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Ping is a simple ping-pong endpoint to check if the server is running properly.
// @Summary      Returns a pong response
// @Description  Simple ping-pong endpoint to check if the server is running properly.
// @Tags         HealthCheck
// @Produce      json
// @Success 200  {object} endpoints.SuccessResponse "Pong response"
// @Router       /v1/ping [get]
func Ping(c *gin.Context) {
	_, span := endpoints.Tracer.Start(c.Request.Context(), "Ping")
	defer span.End()

	// Create a response
	response := endpoints.SuccessResponse{
		Message: "pong",
	}

	endpoints.Logger.Warn(c.Request.Context(), "Pong response sent")

	c.JSON(http.StatusOK, response)
}
