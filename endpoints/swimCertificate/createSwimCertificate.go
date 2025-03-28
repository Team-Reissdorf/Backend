package swimCertificate

import (
	"fmt"
	"net/http"

	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"gorm.io/gorm"
)

// CreateSwimCertificate handles the upload of a swim certificate for an athlete
// @Summary Uploads a swim certificate for an athlete.
// @Description A Trainer can upload a swim certificate as a file for an specified athlete.
// @Tags Swim Certificate
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param AthleteId path int true "Get the latest performance entry of the given athlete_id"
// @Param Authorization header string false "JWT Token"
// @Success 200 {object} endpoints.SuccessResponse "Upload successful"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request"
// @Failure 401 {object} endpoints.ErrorResponse "Unauthorized"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/swimCertificate/create/{AthleteId} [post]
func CreateSwimCertificate(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "CreateSwimCertificate")
	defer span.End()

	//get file from request
	file, errGetFile := c.FormFile("file")
	if errGetFile != nil {
		errGetFile = errors.Wrap(errGetFile, "Failed to retrieve file from request")
		endpoints.Logger.Debug(ctx, errGetFile)
		c.JSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid file upload"})
		return
	}

	// Get the athlete id from the context
	athleteIdString := c.Param("AthleteId")
	if athleteIdString == "" {
		endpoints.Logger.Debug(ctx, "Missing or invalid athlete ID")
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Missing or invalid athlete ID"})
		return
	}
	athleteID, errGetAthleteID := strconv.Atoi(athleteIdString)
	if errGetAthleteID != nil {
		errGetAthleteID = errors.Wrap(errGetAthleteID, "Failed to parse athlete ID")
		endpoints.Logger.Debug(ctx, errGetAthleteID)
		c.AbortWithStatusJSON(http.StatusBadRequest, endpoints.ErrorResponse{Error: "Invalid athlete ID"})
		return
	}

	// Check if the user exists and is assigned to the correct trainer
	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)
	exists, errCheckAthleteTrainer := athleteManagement.AthleteExistsForTrainer(ctx, uint(athleteID), trainerEmail)
	if errCheckAthleteTrainer != nil {
		errCheckAthleteTrainer = errors.Wrap(errCheckAthleteTrainer, "Failed to check if the athlete exists and is assigned to the trainer")
		endpoints.Logger.Error(ctx, errCheckAthleteTrainer)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the athlete exists"})
		return
	}
	if !exists {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Athlete with id %d does not exist", athleteID))
		c.AbortWithStatusJSON(http.StatusNotFound, "Athlete does not exist")
		return
	} else {
		endpoints.Logger.Debug(ctx, fmt.Sprintf("Athlete with id %d exists and is assigned to the given trainer", athleteID))
	}

	//create directory to store files
	uploadDir := filepath.Join("uploads", "swimCertificates", "athlete_"+strconv.Itoa(athleteID))
	if errCreateDir := os.MkdirAll(uploadDir, os.ModePerm); errCreateDir != nil {
		errCreateDir = errors.Wrap(errCreateDir, "Failed to create upload directory")
		endpoints.Logger.Error(ctx, errCreateDir)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Could not create directory"})
		return
	}

	//save uploaded file with unique name
	uniqueFileName := uuid.New().String() + filepath.Ext(file.Filename)
	filePath := filepath.Join(uploadDir, uniqueFileName)

	if errSaveFile := c.SaveUploadedFile(file, filePath); errSaveFile != nil {
		errSaveFile = errors.Wrap(errSaveFile, "Failed to save uploaded file into directory")
		endpoints.Logger.Error(ctx, errSaveFile)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Could not save file"})
		return
	}

	//create swimCertificate object & load in DB
	swimCert := databaseUtils.SwimCertificate{
		AthleteId:    		uint(athleteID),
		Date:         		time.Now(),
		DocumentPath: 		filePath,
		OriginalFileName: 	file.Filename,
	}

	errSaveToDB := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		return tx.Create(&swimCert).Error
	})
	if errSaveToDB != nil { //error if something went wrong with saving to DB
		endpoints.Logger.Error(ctx, errSaveToDB)
		c.JSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Could not save swim certificate"})
		return
	}

	//return succesful response
	c.JSON(http.StatusOK, endpoints.SuccessResponse{Message: "File uploaded successfully"})
}
