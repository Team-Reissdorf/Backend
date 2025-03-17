package athleteManagement

import (
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/formatHelper"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
)

// CreateAthlete creates a new athlete profile
// @Summary Creates a new athlete profile
// @Description Creates a new athlete profile with the given data. Duplicate email and birthdate combinations are not allowed.
// @Tags Athlete Management
// @Accept json
// @Produce json
// @Param Athlete body AthleteBody true "Details of an athlete to create a profile"
// @Param Authorization  header  string  false  "Access JWT is sent in the Authorization header or set as a http-only cookie"
// @Success 201 {object} endpoints.SuccessResponse "Creation successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request body"
// @Failure 401 {object} endpoints.ErrorResponse "The token is invalid"
// @Failure 409 {object} endpoints.ErrorResponse "Athlete already exists"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/athlete/create [post]
func CreateAthlete(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreateAthlete")
	defer span.End()

	// Bind JSON body to struct
	var body AthleteBody
	if err := c.ShouldBindJSON(&body); err != nil {
		err = errors.Wrap(err, "Failed to bind JSON body")
		endpoints.Logger.Debug(ctx, err)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid request body"})
		c.Abort()
		return
	}

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Translate into a database object
	athleteBodies := make([]AthleteBody, 1)
	athleteBodies[0] = body
	athleteEntries := translateAthleteBodies(ctx, athleteBodies, trainerEmail)

	// Validate the athlete body
	err1 := validateAthlete(ctx, &athleteEntries[0])
	if errors.Is(err1, formatHelper.EmptyStringError) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: err1.Error()})
		return
	} else if errors.Is(err1, formatHelper.InvalidSexLengthError) || errors.Is(err1, formatHelper.InvalidSexValue) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Sex needs to be <m|f|d>"})
		return
	} else if errors.Is(err1, formatHelper.DateFormatInvalidError) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid date format"})
		return
	} else if errors.Is(err1, formatHelper.InvalidEmailAddressFormatError) || errors.Is(err1, formatHelper.EmailAddressContainsNameError) || errors.Is(err1, formatHelper.EmailAddressInvalidTldError) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid email address format"})
		return
	} else if errors.Is(err1, formatHelper.DateInFutureError) {
		endpoints.Logger.Debug(ctx, err1)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Date is in the Future"})
		return
	} else if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to validate the athlete body")
		endpoints.Logger.Error(ctx, err1)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Internal server error"})
		return
	}

	// Create the athlete
	err2, alreadyExistingAthletes := createNewAthletes(ctx, athleteEntries)
	if errors.Is(err2, NoNewAthletesError) {
		endpoints.Logger.Debug(ctx, err2)
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: "No new Athletes"})
		return
	} else if err2 != nil {
		err2 = errors.Wrap(err2, "Failed to create the athlete")
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to create the athlete"})
		return
	}

	// Check if the athlete already exists
	if len(alreadyExistingAthletes) > 0 { // Should never happen, since the NoNewAthletesError should be thrown
		err := errors.New("Athlete already exists")
		endpoints.Logger.Error(ctx, err)
		c.AbortWithStatusJSON(http.StatusConflict, endpoints.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(
		http.StatusCreated,
		endpoints.SuccessResponse{
			Message: "Creation successful",
		},
	)
}
