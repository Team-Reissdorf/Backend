package endpoints

import (
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Ping is a simple ping-pong endpoint to check if the server is running properly.
// @Summary      Returns a pong response
// @Description  Simple ping-pong endpoint to check if the server is running properly.
// @Tags         HealthCheck
// @Produce      json
// @Success 200  {object} standardJsonAnswers.SuccessResponse "Pong response"
// @Router       /v1/ping [get]
func Ping(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "Ping")
	defer span.End()

	// Create a response
	response := standardJsonAnswers.SuccessResponse{
		Message: "pong",
	}

	logger.Warn(c.Request.Context(), "Pong response sent")

	c.JSON(http.StatusOK, response)
}

// Teapot is an endpoint that returns a 418 I'm a teapot response.
// @Summary      Returns a 418 I'm a teapot response
// @Description  Simple endpoint to return a 418 I'm a teapot response.
// @Tags         HealthCheck
// @Produce      json
// @Success      418 {object} standardJsonAnswers.ErrorResponse "Error: I'm a teapot"
// @Router       /v1/coffee [get]
func Teapot(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "Teapot")
	defer span.End()

	// Create a response
	response := standardJsonAnswers.ErrorResponse{
		Error: "I'm a teapot. I brew tea, not coffee. ☕🫖",
	}

	logger.Warn(c.Request.Context(), "Teapot response sent")

	c.JSON(http.StatusTeapot, response)
}
