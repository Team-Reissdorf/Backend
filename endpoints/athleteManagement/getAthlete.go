package athleteManagement

import (
	"net/http"
	"strconv"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
// @Router /v1/athlete/get/{AthleteId} [get]
func GetAthleteByID(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "GetOneAthlete")
	defer span.End()

	// Get the athlete id from the context
	athleteIdString := c.Param("AthleteId")
	var athleteId uint
	if athleteIdString == "" {
		endpoints.Logger.Debug(ctx, "Missing or invalid athlete ID")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Missing or invalid athlete ID"})
		return
	} else {
		athleteIdInt, err := strconv.ParseUint(athleteIdString, 10, 32)
		if err != nil {
			err = errors.Wrap(err, "Failed to parse athlete id")
			endpoints.Logger.Debug(ctx, err)
			c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid athlete ID"})
			return
		}
		athleteId = uint(athleteIdInt)
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Get the specified athlete if he corresponds to the given trainer
	athlete, err2 := GetAthlete(ctx, athleteId, trainerEmail)
	if errors.Is(err2, gorm.ErrRecordNotFound) {
		err2 = errors.Wrap(err2, "Athlete not found")
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete not found"})
		return
	} else if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to get the athlete")
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athlete"})
		return
	}

	// get the swim cert for the specific athlete
	var cert databaseUtils.SwimCertificate
	err_swimcert := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		id := athlete.ID

		res := tx.Where("athlete_id = ?", id).First(&cert)
		return errors.Wrap(res.Error, "failed to get swim certificate for athlete")
	})

	var flag bool
	if err_swimcert != nil {
		flag = false
	} else {
		// kind of "silent fail" might want to change
		flag = true
	}

	// Translate athlete to response type
	athleteBody, err3 := translateAthleteToResponse(ctx, *athlete, flag)
	if err3 != nil {
		err3 = errors.Wrap(err3, "Failed to translate the athlete")
		endpoints.Logger.Error(ctx, err3)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		return
	}

	// Send successful response
	c.JSON(
		http.StatusOK,
		AthleteResponse{
			Message: "Request successful",
			Athlete: *athleteBody,
		},
	)
}
