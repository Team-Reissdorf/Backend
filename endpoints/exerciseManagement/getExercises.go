package exerciseManagement

import (
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/endpoints/disciplineManagement"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

type ExercisesResponse struct {
	Message   string               `json:"message" example:"Request successful"`
	Exercises []ExerciseBodyWithId `json:"exercises"`
}

// GetExercisesOfDiscipline returns all exercises of the given discipline. When the athlete id is given, the age specific description will be returned with the exercise.
// @Summary Returns the exercises
// @Description All exercises of the given discipline will be returned. When the athlete id is given, the age specific description will be returned with the exercise.
// @Tags Exercise Management
// @Produce json
// @Param DisciplineName path string true "Get the exercises with the given discipline name"
// @Param athlete-id query uint false "Get the exercise_specifics for the given athletes age"
// @Param performance-date query string false "Date in YYYY-MM-DD format to get the exercises according to the ruleset of the given year"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} ExercisesResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Discipline, athlete or ruleset year does not exist"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/exercise/get/{DisciplineName} [get]
func GetExercisesOfDiscipline(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetExercisesOfDiscipline")
	defer span.End()

	// Get the athlete id from the context
	disciplineName := c.Param("DisciplineName")
	if disciplineName == "" {
		endpoints.Logger.Debug(ctx, "Missing or invalid discipline name")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Missing or invalid discipline name"})
		return
	}

	// Get the athlete_id query parameter from the context
	athleteIdString := c.Query("athlete-id")
	athleteIdIsSet := athleteIdString != ""
	var athleteId uint
	if athleteIdIsSet {
		athleteIdInt, err := strconv.ParseUint(athleteIdString, 10, 32)
		if err != nil {
			err = errors.Wrap(err, "Failed to parse 'athlete-id' query parameter")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'athlete-id' query parameter"})
			return
		}
		athleteId = uint(athleteIdInt)
	}

	// Get the performance year from the context
	performanceDateString := c.Query("performance-date")
	var performanceYear int
	performanceDateIsSet := performanceDateString != ""
	if performanceDateIsSet {
		// Parse the performance date
		t, err1 := time.Parse(time.DateOnly, performanceDateString)
		if err1 != nil {
			err1 = errors.Wrap(err1, "Failed to parse date: "+performanceDateString)
			endpoints.Logger.Debug(ctx, err1)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'performance-date' query parameter"})
			return
		}
		performanceYear = t.Year()

		// Check if date is in the past
		if err := formatHelper.IsFuture(performanceDateString); err != nil {
			err = errors.Wrap(err, "'performance-date' is in the future")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "'performance-date' is in the Future"})
			return
		}
	}

	// Check if a ruleset for the given year exists
	if performanceDateIsSet {
		var rulesetCount int64
		err := DatabaseFlow.GetDB(ctx).
			Model(&databaseUtils.Ruleset{}).
			Where("year = ?", strconv.Itoa(performanceYear)).
			Count(&rulesetCount).
			Error
		if err != nil {
			err = errors.Wrap(err, "Failed to get rulesets")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Ruleset for the given performance year does not exist"})
			return
		}
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Check if the given discipline exists
	disciplineExists, err1 := disciplineManagement.DisciplineExists(ctx, disciplineName)
	if err1 != nil {
		endpoints.Logger.Error(ctx, err1)
		// Move on since the discipline could exist and the following request can handle not existing disciplines
	}
	if !disciplineExists {
		endpoints.Logger.Debug(ctx, "Discipline does not exist")
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Discipline does not exist"})
		return
	}

	// Get the athletes age
	var age int
	if athleteIdIsSet {
		athlete, errA := athleteManagement.GetAthlete(ctx, athleteId, trainerEmail)
		// Check if the athlete could be found
		if errors.Is(errA, gorm.ErrRecordNotFound) {
			err := errors.New("Athlete does not exist")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: err.Error()})
			return
		} else if errA != nil {
			errA = errors.Wrap(errA, "Failed to get athlete")
			endpoints.Logger.Debug(ctx, errA)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete"})
			return
		}

		// Calculate the age of the athlete
		birthDate, errB := formatHelper.FormatDate(athlete.BirthDate)
		if errB != nil {
			errB = errors.Wrap(errB, "Failed to parse the birth date")
			endpoints.Logger.Error(ctx, errB)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to parse the birth date"})
			return
		}
		var errC error
		age, errC = athleteManagement.CalculateAge(ctx, birthDate)
		if errC != nil {
			errC = errors.Wrap(errC, "Failed to calculate the age of the athlete")
			endpoints.Logger.Error(ctx, errC)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete's age"})
			return
		}
	}

	// Get the exercises, and optionally filter for the age and ruleset year
	var results []ExerciseBodyWithId
	db := DatabaseFlow.GetDB(ctx)
	query := db.Model(&databaseUtils.Exercise{})

	if athleteIdIsSet {
		query = query.
			Joins("JOIN exercise_rulesets ON exercise_rulesets.exercise_id = exercises.id").
			Joins("JOIN exercise_goals ON exercise_goals.ruleset_id = exercise_rulesets.id AND exercise_goals.from_age <= ? AND exercise_goals.to_age >= ?", age, age).
			Select("exercises.id as exercise_id, exercises.name, exercises.unit, exercises.discipline_name, exercise_goals.description as age_specifics")
	} else {
		query = query.Select("exercises.id as exercise_id, exercises.name, exercises.unit, exercises.discipline_name")
	}

	if performanceDateIsSet && athleteIdIsSet {
		query = query.
			Where("exercise_rulesets.ruleset_year = ?", strconv.Itoa(performanceYear))
	} else if performanceDateIsSet {
		query = query.
			Joins("JOIN exercise_rulesets ON exercise_rulesets.exercise_id = exercises.id").
			Where("exercise_rulesets.ruleset_year = ?", strconv.Itoa(performanceYear))
	}

	err2 := query.
		Where("discipline_name = ?", disciplineName).
		Find(&results).
		Error
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to get exercises")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get exercises"})
		return
	}

	c.JSON(
		http.StatusOK,
		ExercisesResponse{
			Message:   "Request successful",
			Exercises: results,
		},
	)
}
