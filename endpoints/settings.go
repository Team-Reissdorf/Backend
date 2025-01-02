package endpoints

import (
	"fmt"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/Team-Reissdorf/Backend/endpoints/standardJsonAnswers"
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
// @Success 200 {object} standardJsonAnswers.SuccessResponse "Log level changed"
// @Failure 400 {object} standardJsonAnswers.ErrorResponse "Invalid request body"
// @Failure 401 {object} standardJsonAnswers.ErrorResponse "The token is invalid"
// @Failure 500 {object} standardJsonAnswers.ErrorResponse "Internal server error"
// @Router /v1/settings/change-log-level [post]
func ChangeLogLevel(c *gin.Context) {
	ctx, span := tracer.Start(c.Request.Context(), "Change log level")
	defer span.End()

	// Get the new log level from the request
	var logLevelHolder ChangeLogLevelRequest
	if err := c.BindJSON(&logLevelHolder); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body for log level change")
		logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: "Invalid request body"})
		c.Abort()
		return
	}

	// Check if the log level is valid
	if logLevelHolder.LogLevel < 0 || logLevelHolder.LogLevel > 4 {
		err := errors.New("Invalid log level")
		logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, standardJsonAnswers.ErrorResponse{Error: "Invalid log level"})
		c.Abort()
		return
	}

	// Change the log level
	lvl := FlowWatch.Level(logLevelHolder.LogLevel)
	FlowWatch.SetLogLevel(lvl)
	msg := fmt.Sprintf("Log level changed to: %s", lvl.String())
	logger.Warn(ctx, msg)
	c.JSON(
		http.StatusOK,
		standardJsonAnswers.SuccessResponse{
			Message: msg,
		},
	)
}
