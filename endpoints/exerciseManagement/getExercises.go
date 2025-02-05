package exerciseManagement

import (
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/endpoints/disciplineManagement"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
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
// @Param athlete_id query uint false "Get the exercise_specifics for the given athletes age"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} ExercisesResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Discipline or athlete does not exist"
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
	athleteIdString := c.Query("athlete_id")
	athleteIdIsSet := athleteIdString != ""
	var athleteId uint
	if athleteIdIsSet {
		athleteIdInt, err := strconv.ParseUint(athleteIdString, 10, 32)
		if err != nil {
			err = errors.Wrap(err, "Failed to parse 'athlete_id' query parameter")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid 'athlete_id' query parameter"})
			return
		}
		athleteId = uint(athleteIdInt)
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

	// Check if the athlete exists if the id is given
	if athleteIdIsSet {
		// Check if the athlete exists for the given trainer
		athleteExists, err := athleteManagement.AthleteExistsForTrainer(ctx, athleteId, trainerEmail)
		if err != nil {
			endpoints.Logger.Error(ctx, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the athlete exists"})
			return
		}
		if !athleteExists {
			endpoints.Logger.Debug(ctx, "Athlete does not exist")
			c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete does not exist"})
			return
		}
	}

	// Get all exercises of the discipline from the database
	var exercises []databaseUtils.Exercise
	err2 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Model(databaseUtils.Exercise{}).Where("discipline_name = ?", disciplineName).Find(&exercises).Error
		return err
	})
	if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to get all exercises of the discipline")
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the exercises of the discipline"})
		return
	}

	// Translate exercises to response type
	exercisesResponse := make([]ExerciseBodyWithId, len(exercises))
	for idx, exercise := range exercises {
		exerciseBody, err3 := translateExerciseToResponse(ctx, exercise)
		if err3 != nil {
			err3 = errors.Wrap(err3, "Failed to translate the exercise")
			endpoints.Logger.Error(ctx, err3)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
			return
		}

		exercisesResponse[idx] = *exerciseBody
	}

	c.JSON(
		http.StatusOK,
		ExercisesResponse{
			Message:   "Request successful",
			Exercises: exercisesResponse,
		},
	)
}
