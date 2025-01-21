package backendSettings

import (
	"fmt"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

type ChangeLogLevelRequest struct {
	LogLevel int `json:"log_level" example:"1"`
}

// ChangeLogLevel changes the log level of the server
// @Summary ChangeLogLevel changes the log level of the server
// @Description Changes the log level of the server to the specified level 0-4 (Debug-Fatal).
// @Tags Settings
// @Accept json
// @Produce json
// @Param Authorization  header  string  false  "Settings access JWT is sent in the Authorization header or set as a http-only cookie"
// @Param Log-level body ChangeLogLevelRequest true "Log level"
// @Success 200 {object} endpoints.SuccessResponse "Log level changed"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/backendSettings/change-log-level [post]
func ChangeLogLevel(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "Change log level")
	defer span.End()

	// Get the new log level from the request
	var logLevelHolder ChangeLogLevelRequest
	if err := c.BindJSON(&logLevelHolder); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body for log level change")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		c.Abort()
		return
	}

	// Check if the log level is valid
	if logLevelHolder.LogLevel < 0 || logLevelHolder.LogLevel > 4 {
		err := errors.New("Invalid log level")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid log level"})
		c.Abort()
		return
	}

	// Change the log level
	lvl := FlowWatch.Level(logLevelHolder.LogLevel)
	FlowWatch.SetLogLevel(lvl)
	msg := fmt.Sprintf("Log level changed to: %s", lvl.String())
	endpoints.Logger.Warn(ctx, msg)
	c.JSON(
		http.StatusOK,
		endpoints.SuccessResponse{
			Message: msg,
		},
	)
}
