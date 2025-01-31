package athleteManagement

import (
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseModels"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

// DeleteAthlete deletes the given athlete profile
// @Summary Deletes the given athlete profile
// @Description Deletes the given athlete profile.
// @Tags Athlete Management
// @Produce json
// @Param AthleteId path int true "Delete the given athlete"
// @Param Authorization  path  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 200 {object} endpoints.SuccessResponse "Deletion successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request parameter"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 404 {object} endpoints.ErrorResponse "Athlete could not be found for this trainer"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/delete/{AthleteId} [delete]
func DeleteAthlete(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "DeleteAthlete")
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

	// Delete the athlete from the database
	err2 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		result := tx.Delete(&databaseModels.Athlete{}, "trainer_email = ? AND athlete_id = ?", trainerEmail, athleteId)
		if result.Error != nil {
			return errors.Wrap(result.Error, "Failed to delete the athlete")
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if errors.Is(err2, gorm.ErrRecordNotFound) {
		err2 = errors.Wrap(err2, "Athlete not found")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete not found"})
		return
	} else if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to delete the athlete"})
		return
	}

	c.JSON(http.StatusOK, endpoints.SuccessResponse{Message: "Deletion successful"})
}
