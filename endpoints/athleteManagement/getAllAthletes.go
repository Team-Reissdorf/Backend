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
	Message  string              `json:"message" example:"Request successful"`
	Athletes []AthleteBodyWithId `json:"athletes"`
	// SwimCerts []SwimCertificateWithID `json:"swimcerts"`
}

// GetAllAthletes returns all athletes
// @Summary Returns all athlete profiles and swim certificates
// @Description All athlete profiles and swim certificates of the given trainer are returned
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

	// get all swim certs
	// one could also create an sql query that only gets the swim certs related to the trainer
	var certs []databaseUtils.SwimCertificate

	err2 := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {

		// ids := make([]uint, len(athletes))
		// for idx, athlete := range athletes {
		// 	ids[idx] = athlete.ID
		// }

		res := tx.Find(&certs)
		// Where("AthleteID IN ?", ids)

		return errors.Wrap(res.Error, "failed to get swim certificates")
	})

	if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to get the swim certs"})
		c.Abort()
		return
	}

	// convert into return type
	// cert_r := make([]SwimCertificateWithID, len(certs))
	// var i int

	// for i = 0; i < len(certs); i++ {
	// 	cert := certs[i]

	// 	// make sure that the swim cert applies for an athlete created by the trainer
	// 	// var i2 int
	// 	check := false
	// 	for idx, _ := range athletes {
	// 		if cert.AthleteId == athletes[idx].ID {
	// 			check = true
	// 			break
	// 		}
	// 	}

	// 	if !check {
	// 		continue
	// 	} else {
	// 		cert_r[i] = SwimCertificateWithID{
	// 			ID:        cert.ID,
	// 			AthleteId: cert.AthleteId,
	// 		}
	// 	}
	// 	// cert_r[i] = SwimCertificateWithID{
	// 	// 	ID:        cert.ID,
	// 	// 	AthleteId: cert.AthleteId,
	// 	// }

	// }

	// Translate athletes to response type
	athletesResponse := make([]AthleteBodyWithId, len(athletes))
	for idx, athlete := range athletes {
		// Translate athlete to response type
		sc_flag := false
		for _, cert := range certs {
			if cert.AthleteId == athlete.ID {
				sc_flag = true
			}
		}
		athleteBody, err2 := translateAthleteToResponse(ctx, athlete, sc_flag)
		if err2 != nil {
			err2 = errors.Wrap(err2, "Failed to translate the athlete")
			endpoints.Logger.Error(ctx, err2)
			c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
			return
		}

		athletesResponse[idx] = *athleteBody
	}

	// Send successful response
	c.JSON(
		http.StatusOK,
		AthletesResponse{
			Message:  "Request successful",
			Athletes: athletesResponse,
			// SwimCerts: cert_r,
		},
	)
}
