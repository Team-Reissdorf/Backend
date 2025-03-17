package performanceManagement

import (
	"fmt"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
)

// EditPerformanceEntry edits a performance entry
// @Summary Edits an existing performance entry
// @Description Edits an existing performance entry with the given details.
// @Tags Performance Management
// @Produce json
// @Param Performance body PerformanceBodyEdit true "Edited details of a performance entry"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} endpoints.SuccessResponse "Edited successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Performance entry or goals not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/edit [put]
func EditPerformanceEntry(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "EditPerformanceEntry")
	defer span.End()

	// Bind JSON body to struct
	var body PerformanceBodyEdit
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	if err1 := formatHelper.IsEmpty(body.Date); err1 != nil {
		endpoints.Logger.Debug(ctx, err1)
		err1 = errors.Wrap(err1, "Date is empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err1.Error()})
		return
	}

	// Check if the given performance entry is for an athlete of the given trainer
	exists, err1 := performanceExistsForTrainer(ctx, body.PerformanceId, trainerEmail)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to check if the performance entry exists and is assigned to the trainer")
		endpoints.Logger.Error(ctx, err1)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the performance entry exists"})
		return
	}
	if !exists {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Performance entry with id %d does not exist", body.PerformanceId))
		c.AbortWithStatusJSON(http.StatusNotFound, "Performance entry does not exist")
		return
	} else {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Performance entry with id %d exists and is assigned to the given trainer", body.PerformanceId))
	}

	// Get the athlete for the given trainer
	athlete, err2 := athleteManagement.GetAthleteFromPerformanceId(ctx, body.PerformanceId, trainerEmail)
	if errors.Is(err2, gorm.ErrRecordNotFound) {
		err2 = errors.Wrap(err2, "Athlete does not exist")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete does not exist"})
		return
	} else if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to get the athlete")
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete"})
		return
	}

	// Check if the creation limit is reached
	count, err3 := countPerformanceEntriesPerDisciplinePerDayEditMode(ctx, athlete.ID, body.ExerciseId, body.PerformanceId, body.Date)
	if err3 != nil {
		endpoints.Logger.Error(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to create the performance entry"})
		return
	}
	if uint8(count) >= limitPerDisciplinePerDay {
		err := errors.New("The athlete has reached the daily limit for this discipline")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: err.Error()})
		return
	}

	// Calculate the age of the athlete
	birthDate, err4 := formatHelper.FormatDate(athlete.BirthDate)
	if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to parse the birth date")
		endpoints.Logger.Error(ctx, err4)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to parse the birth date"})
		return
	}
	age, err5 := athleteManagement.CalculateAge(ctx, birthDate)
	if err5 != nil {
		err5 = errors.Wrap(err5, "Failed to calculate the age of the athlete")
		endpoints.Logger.Error(ctx, err5)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete's age"})
		return
	}

	// Get the corresponding medal status
	medal, err6 := evaluateMedalStatus(ctx, body.ExerciseId, age, athlete.Sex, body.Points)
	if errors.Is(err6, gorm.ErrRecordNotFound) {
		err6 = errors.Wrap(err6, "No exercise goals for this athlete found")
		endpoints.Logger.Debug(ctx, err6)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "No exercise goals found for this athlete"})
		return
	} else if err6 != nil {
		err6 = errors.Wrap(err6, "Failed to calculate the medal status")
		endpoints.Logger.Error(ctx, err6)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the goals for this athlete"})
		return
	}

	//Check if the Date is in Correct Format
	err7 := formatHelper.IsDate(body.Date)
	if errors.Is(err7, formatHelper.DateFormatInvalidError) {
		endpoints.Logger.Debug(ctx, err7)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
		return
	}

	//Check if the Date is in the Past
	err8 := formatHelper.IsFuture(body.Date)
	if errors.Is(err8, formatHelper.DateInFutureError) {
		err8 = errors.Wrap(err8, "Date is in the future")
		endpoints.Logger.Debug(ctx, err8)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Date is in the future"})
		return
	} else if err8 != nil {
		err8 = errors.Wrap(err8, "Failed to check the date")
		endpoints.Logger.Error(ctx, err8)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check the date"})
		return
	}

	// Translate to database entry
	performanceEntry := databaseUtils.Performance{
		ID:         body.PerformanceId,
		Points:     body.Points,
		Date:       body.Date,
		ExerciseId: body.ExerciseId,
		Medal:      medal,
	}

	// Update the performance entry in the database
	err9 := updatePerformanceEntry(ctx, performanceEntry)
	if err9 != nil {
		err9 = errors.Wrap(err9, "Failed to update the performance entry")
		endpoints.Logger.Error(ctx, err9)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to update the performance entry"})
		return
	}

	c.JSON(
		http.StatusOK,
		endpoints.SuccessResponse{
			Message: "Edited successful",
		},
	)
}
