package athleteManagement

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type AthletesResponse struct {
	Message  string              `json:"message" example:"Request successful"`
	Athletes []AthleteBodyWithId `json:"athletes"`
}

// GetAllAthletes returns all athletes
// @Summary Returns all athlete profiles
// @Description All athlete profiles of the given trainer are returned
// @Tags Athlete Management
// @Produce json
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} AthletesResponse "Request successful"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/get-all [get]
func GetAllAthletes(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetAllAthletes")
	defer span.End()

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user is a trainer
	trainerEmail := "blabla@test.com"

	// Get all athletes for the given trainer
	var athletes []databaseModels.Athlete
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Where("trainer_email LIKE ?", strings.ToLower(trainerEmail)).Find(&athletes).Error
		err = errors.Wrap(err, "Failed to get the athletes")
		return err
	})
	if err1 != nil {
		endpoints.Logger.Error(ctx, err1)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athletes"})
		c.Abort()
		return
	}

	c.JSON(
		http.StatusOK,
		AthletesResponse{
			Message:  "Request successful",
			Athletes: translateAthletesToResponse(ctx, athletes),
		},
	)
}

// translateAthletesToResponse converts athlete database objects to response type
func translateAthletesToResponse(ctx context.Context, athletes []databaseModels.Athlete) []AthleteBodyWithId {
	ctx, span := endpoints.Tracer.Start(ctx, "TranslateAthletesToResponse")
	defer span.End()

	athletesResponse := make([]AthleteBodyWithId, len(athletes))
	for idx, athlete := range athletes {
		athletesResponse[idx] = AthleteBodyWithId{
			AthleteId: athlete.AthleteId,
			FirstName: athlete.FirstName,
			LastName:  athlete.LastName,
			Email:     athlete.Email,
			BirthDate: athlete.BirthDate,
			Sex:       athlete.Sex,
		}
	}

	return athletesResponse
}
