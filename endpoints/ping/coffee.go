package ping

import (
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Teapot is an endpoint that returns a 418 I'm a teapot response.
// @Summary      Returns a 418 I'm a teapot response
// @Description  Simple endpoint to return a 418 I'm a teapot response.
// @Tags         HealthCheck
// @Produce      json
// @Success      418 {object} endpoints.ErrorResponse "Error: I'm a teapot"
// @Router       /v1/coffee [get]
func Teapot(c *gin.Context) {
	_, span := endpoints.Tracer.Start(c.Request.Context(), "Teapot")
	defer span.End()

	// Create a response
	response := endpoints.ErrorResponse{
		Error: "I'm a teapot. I brew tea, not coffee. ☕🫖",
	}

	endpoints.Logger.Warn(c.Request.Context(), "Teapot response sent")

	c.JSON(http.StatusTeapot, response)
}
