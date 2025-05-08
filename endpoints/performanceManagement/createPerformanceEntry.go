package performanceManagement

import (
	//"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type CreatePerformanceResponse struct {
	Message     string `json:"message" example:"Creation successful"`
	MedalStatus string `json:"medal_status" example:"Medal status"`
}

var limitPerDisciplinePerDay uint8 = 3

// CreatePerformance creates a new performance entry
// @Summary Creates a new performance entry
// @Description Creates a new performance entry with the given data. Maximum 3 entries/discipline/athlete/day are allowed (performance limit).
// @Tags Performance Management
// @Accept json
// @Produce json
// @Param Performance body PerformanceBody true "Details of a performance (valid units are: <millisecond, centimeter, point, bool>)"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} CreatePerformanceResponse "Creation successful"
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
	if err1 := formatHelper.IsEmpty(body.Date); err1 != nil {
		endpoints.Logger.Debug(ctx, err1)
		err1 = errors.Wrap(err1, "Date is empty")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err1.Error()})
		return
	} else if err1 := formatHelper.IsDate(body.Date); err1 != nil {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
		return
	}

	// Check if the date is in past
	err2 := formatHelper.IsFuture(body.Date)
	if errors.Is(err2, formatHelper.DateInFutureError) {
		err2 = errors.Wrap(err2, "Date is in the future")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Date is in the future"})
		return
	} else if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to check the date")
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check the date"})
		return
	}

	// Get the athlete for the given trainer
	athlete, err3 := athleteManagement.GetAthlete(ctx, body.AthleteId, trainerEmail)
	if errors.Is(err3, gorm.ErrRecordNotFound) {
		err3 = errors.Wrap(err3, "Athlete does not exist")
		endpoints.Logger.Debug(ctx, err3)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete does not exist"})
		return
	} else if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to get the athlete")
		endpoints.Logger.Error(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete"})
		return
	}

	// Check if the creation limit is reached
	count, err4 := countPerformanceEntriesPerDisciplinePerDay(ctx, athlete.ID, body.ExerciseId, body.Date)
	if err4 != nil {
		endpoints.Logger.Error(ctx, err4)
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
	birthDate, err5 := formatHelper.FormatDate(athlete.BirthDate)
	if err5 != nil {
		err5 = errors.Wrap(err5, "Failed to parse the birth date")
		endpoints.Logger.Error(ctx, err5)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to parse the birth date"})
		return
	}
	age, err6 := athleteManagement.CalculateAge(ctx, birthDate)
	if err6 != nil {
		err6 = errors.Wrap(err6, "Failed to calculate the age of the athlete")
		endpoints.Logger.Error(ctx, err6)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete's age"})
		return
	}
	// Check if the exercise goal exists for the athlete's age
	exists, err7 := exerciseGoalExistsForAge(ctx, body.ExerciseId, age, strconv.Itoa(time.Now().Year()))
	if err7 != nil {
		endpoints.Logger.Error(ctx, err7)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check exercise goal"})
		return
	}
	if !exists {
		err8 := errors.New("No exercise goal found for the athlete's age")
		endpoints.Logger.Debug(ctx, err8)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "No exercise goal found for the athlete's age"})
		return
	}

	// Translate the performance body to a database entry
	performanceBodies := make([]PerformanceBody, 1)
	performanceBodies[0] = body
	performanceEntries, err9 := translatePerformanceBodies(ctx, performanceBodies, age, athlete.Sex)
	if errors.Is(err9, gorm.ErrRecordNotFound) {
		err9 = errors.Wrap(err9, "No exercise goals for this athlete found")
		endpoints.Logger.Debug(ctx, err9)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "No exercise goals found for this athlete"})
		return
	} else if err9 != nil {
		endpoints.Logger.Error(ctx, err9)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the goals for this athlete"})
		return
	}

	// Create performance entry in the database
	err10 := createNewPerformances(ctx, performanceEntries)
	if errors.Is(err10, databaseUtils.ErrForeignKeyViolation) {
		err10 = errors.Wrap(err10, "Athlete or exercise does not exist")
		endpoints.Logger.Debug(ctx, err10)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete or exercise does not exist"})
		return
	}
	if err10 != nil {
		err10 = errors.Wrap(err10, "Failed to create the performance entry")
		endpoints.Logger.Error(ctx, err10)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to create the performance entry"})
		return
	}

	c.JSON(
		http.StatusCreated,
		CreatePerformanceResponse{
			Message:     "Creation successful",
			MedalStatus: performanceEntries[0].Medal,
		},
	)
}
