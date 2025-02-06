package performanceManagement

import (
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

// CreatePerformance creates a new performance entry
// @Summary Creates a new performance entry
// @Description Creates a new performance entry with the given data. Maximum 3 entries/discipline/athlete/day are allowed (performance limit).
// @Tags Performance Management
// @Accept json
// @Produce json
// @Param Performance body PerformanceBody true "Details of a performance (points should be given in seconds, centimeters, points or as a boolean)"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete or exercise does not exist"
// @Failure 409 {object} endpoints.ErrorResponse "Performance limit reached"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/performance/create [post]
func CreatePerformance(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreatePerformance")
	defer span.End()

	// Bind JSON body to struct
	var body PerformanceBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Validate the date format
	err1 := formatHelper.IsDate(body.Date)
	if err1 != nil {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
		return
	}

	// Get the athlete for the given trainer
	athlete, err2 := athleteManagement.GetAthlete(ctx, body.AthleteId, trainerEmail)
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

	// Calculate the age of the athlete
	birthDate, err3 := formatHelper.FormatDate(athlete.BirthDate)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to parse the birth date")
		endpoints.Logger.Error(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to parse the birth date"})
		return
	}
	age, err4 := athleteManagement.CalculateAge(ctx, birthDate)
	if err4 != nil {
		err4 = errors.Wrap(err4, "Failed to calculate the age of the athlete")
		endpoints.Logger.Error(ctx, err4)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete's age"})
		return
	}

	// Translate the performance body to a database entry
	performanceBodies := make([]PerformanceBody, 1)
	performanceBodies[0] = body
	performanceEntries, err5 := translatePerformanceBodies(ctx, performanceBodies, age, athlete.Sex)
	if errors.Is(err5, gorm.ErrRecordNotFound) {
		err5 = errors.Wrap(err5, "No exercise goals for this athlete found")
		endpoints.Logger.Debug(ctx, err5)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "No exercise goals found for this athlete"})
		return
	} else if err5 != nil {
		endpoints.Logger.Error(ctx, err5)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete's goals for this athlete"})
		return
	}

	// Create performance entry in the database
	err6 := createNewPerformances(ctx, performanceEntries)
	if errors.Is(err6, databaseUtils.ErrForeignKeyViolation) {
		err6 = errors.Wrap(err6, "Athlete or exercise does not exist")
		endpoints.Logger.Debug(ctx, err6)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete or exercise does not exist"})
		return
	}
	if err6 != nil {
		err6 = errors.Wrap(err6, "Failed to create the performance entry")
		endpoints.Logger.Error(ctx, err6)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to create the performance entry"})
		return
	}

	c.JSON(
		http.StatusCreated,
		endpoints.SuccessResponse{
			Message: "Creation successful",
		},
	)
}
