package athleteManagement

import (
	"net/http"
	"strings"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type AthletesResponse struct {
	Message   string                  `json:"message" example:"Request successful"`
	Athletes  []AthleteBodyWithId     `json:"athletes"`
	SwimCerts []SwimCertificateWithID `json:"swimcerts"`
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
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Get all athletes for the given trainer
	var athletes []databaseUtils.Athlete
	err1 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Where("trainer_email = ?", strings.ToLower(trainerEmail)).Find(&athletes).Error
		err = errors.Wrap(err, "Failed to get the athletes")
		return err
	})
	if err1 != nil {
		endpoints.Logger.Error(ctx, err1)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the athletes"})
		c.Abort()
		return
	}

	// get all swim certs for trainer
	var certs []databaseUtils.SwimCertificate
	err2 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		err := tx.Where("trainer_email = ?", strings.ToLower(trainerEmail)).Find(&certs).Error
		err = errors.Wrap(err, "Failed to get the certs")
		return err
	})
	if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the swim certs"})
		c.Abort()
		return
	}

	// Translate athletes to response type
	athletesResponse := make([]AthleteBodyWithId, len(athletes))
	for idx, athlete := range athletes {
		// Translate athlete to response type
		athleteBody, err2 := translateAthleteToResponse(ctx, athlete)
		if err2 != nil {
			err2 = errors.Wrap(err2, "Failed to translate the athlete")
			endpoints.Logger.Error(ctx, err2)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
			return
		}

		athletesResponse[idx] = *athleteBody
	}

	// translate certs to response array
	cert_r := make([]SwimCertificateWithID, len(certs))
	for idx, cert := range certs {
		cert_t := SwimCertificateWithID{
			ID:        cert.ID,
			AthleteId: cert.Athlete.ID,
		}
		cert_r[idx] = cert_t
	}

	// Send successful response
	c.JSON(
		http.StatusOK,
		AthletesResponse{
			Message:   "Request successful",
			Athletes:  athletesResponse,
			SwimCerts: cert_r,
		},
	)
}
