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

type AthleteResponse struct {
	Message string            `json:"message" example:"Request successful"`
	Athlete AthleteBodyWithId `json:"athletes"`
}

func GetAthleteByID(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetOneAthlete")
	defer span.End()

	// Get the athlete id from the context
	athleteID := c.Param("id")

	// Get the user id from the context
	// userId := authHelper.GetUserIdFromContext(ctx, c)
	// ToDo: Verify that the user is a trainer
	trainerEmail := "blabla@test.com"

	var athlete databaseModels.Athlete
	err := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		return tx.Where("trainer_email = ? AND athlete_id = ?", strings.ToLower(trainerEmail), athleteID).
			First(&athlete).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete not found"})
			return
		}
		endpoints.Logger.Error(ctx, err)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete"})
		return
	}

	// Send successful response
	c.JSON(http.StatusOK, translateAthleteToResponse(ctx, athlete))
}

// translateAthleteToResponse converts athlete database objects to response type
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
