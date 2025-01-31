package athleteManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

type AthleteResponse struct {
	Message string            `json:"message" example:"Request successful"`
	Athlete AthleteBodyWithId `json:"athlete"`
}

// GetAthleteByID returns one athlete
// @Summary Returns one athlete profile
// @Description One athlete profile with given id and of the given trainer gets returned
// @Tags Athlete Management
// @Produce json
// @Param AthleteId path int true "Get the athlete with the given id"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} AthletesResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/get-one/{AthleteId} [get]
func GetAthleteByID(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetOneAthlete")
	defer span.End()

	// Get the athlete id from the context
	athleteIdString := c.Param("AthleteId")
	if athleteIdString == "" {
		endpoints.Logger.Debug(ctx, "Missing or invalid athlete ID")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Missing or invalid athlete ID"})
		return
	}
	athleteId, err1 := strconv.Atoi(athleteIdString)
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse athlete ID")
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid athlete ID"})
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Get the specified athlete if he corresponds to the given trainer
	var athlete databaseModels.Athlete
	err := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		return tx.Where("trainer_email = ? AND athlete_id = ?", strings.ToLower(trainerEmail), athleteID).
			First(&athlete).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete not found"})
		c.Abort()
		return
	} else if err != nil {
		endpoints.Logger.Error(ctx, err)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete"})
		c.Abort()
		return
	}

	// Send successful response
	c.JSON(
		http.StatusOK,
		AthleteResponse{
			Message: "Request successful",
			Athlete: translateAthleteToResponse(ctx, athlete),
		},
	)
}

// translateAthleteToResponse converts an athlete database object to response type
func translateAthleteToResponse(ctx context.Context, athlete databaseModels.Athlete) AthleteBodyWithId {
	ctx, span := endpoints.Tracer.Start(ctx, "TranslateAthleteToResponse")
	defer span.End()

	athleteResponse := AthleteBodyWithId{
		AthleteId: athlete.AthleteId,
		FirstName: athlete.FirstName,
		LastName:  athlete.LastName,
		Email:     athlete.Email,
		BirthDate: athlete.BirthDate,
		Sex:       athlete.Sex,
	}

	return athleteResponse
}
