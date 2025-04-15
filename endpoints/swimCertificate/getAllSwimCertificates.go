package swimCertificate

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"strconv"	
	"path/filepath"

	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// DownloadAllSwimCertificates gets all swimcerticates of the athlete and returns them as a ZIP file
// @Summary Download all swim certificates for an athlete
// @Description A trainer can download all swim certificates for a specified athlete. The result is a ZIP file containing all documents.
// @Tags Swim Certificate
// @Produce application/zip
// @Param AthleteId path int true "ID of the athlete"
// @Param Authorization header string false "JWT token"
// @Success 200 {file} file "ZIP archive with all swim certificates"
// @Failure 400 {object} endpoints.ErrorResponse "Invalid request"
// @Failure 401 {object} endpoints.ErrorResponse "Unauthorized"
// @Failure 404 {object} endpoints.ErrorResponse "Swim certificates not found or athlete not found"
// @Failure 500 {object} endpoints.ErrorResponse "Internal server error"
// @Router /v1/swimCertificate/download-all/{AthleteId} [get]
func DownloadAllSwimCertificates(c *gin.Context) {
	ctx, span := endpoints.Tracer.Start(c.Request.Context(), "DownloadAllSwimCertificates")
	defer span.End()

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

	// Get the user id from the context
	trainerEmail := authHelper.GetUserIdFromContext(ctx, c)

	// Check if the athlete exists for the given trainer
	exists, err2 := athleteManagement.AthleteExistsForTrainer(ctx, uint(athleteID), trainerEmail)
	if err2 != nil {
		endpoints.Logger.Error(ctx, err2)
		c.AbortWithStatusJSON(http.StatusInternalServerError, endpoints.ErrorResponse{Error: "Failed to check if the athlete exists"})
		return
	}
	if !exists {
		endpoints.Logger.Debug(ctx, "Athlete does not exist")
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Athlete not found"})
		return
	}

	//get swim certificates from DB
	var certificates []databaseUtils.SwimCertificate
	errGetSCFromDatabase := DatabaseFlow.TransactionHandler(ctx, func(tx *gorm.DB) error {
		return tx.Model(&databaseUtils.SwimCertificate{}).
			Select("document_path, original_file_name").
			Where("athlete_id = ?", athleteID).
			Order("date DESC").
			Find(&certificates).Error
	})
	if errGetSCFromDatabase != nil{
		errGetSCFromDatabase = errors.Wrap(errGetSCFromDatabase, "Failed to get swim certificates")
		endpoints.Logger.Error(ctx, errGetSCFromDatabase)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "Failed to get swim certificates"})
		return
	}
	if len(certificates) == 0 {
		endpoints.Logger.Debug(ctx, "No swim certificates found for athlete with ID: " + athleteIdString)
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.ErrorResponse{Error: "No swim certificates found for this athlete"})
		return
	}
	

	// create ZIP for response 
	c.Writer.Header().Set("Content-Type", "application/zip")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=swim_certificates.zip")

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	// copy swim certificates to ZIP & rename to original 
	usedNames := make(map[string]int) // map to find duplicates  
	for _, cert := range certificates {
		originalName := cert.OriginalFileName
		baseName := originalName
		counter := 1

		// change file name if duplicate name: file.txt -> file (1).txt
		for {
			if _, exists := usedNames[baseName]; !exists {
				break
			}
			ext := filepath.Ext(originalName)
			nameOnly := originalName[:len(originalName)-len(ext)]
			baseName = nameOnly + " (" + strconv.Itoa(counter) + ")" + ext
			counter++
		}
		usedNames[baseName] = 1

		file, err := os.Open(cert.DocumentPath)
		if err != nil {
			endpoints.Logger.Warn(ctx, "Datei konnte nicht geöffnet werden: "+cert.DocumentPath)
			continue
		}
		defer file.Close()

		fw, err := zipWriter.Create(baseName)
		if err != nil {
			endpoints.Logger.Warn(ctx, "ZIP-Eintrag konnte nicht erstellt werden für: "+baseName)
			continue
		}
		if _, err = io.Copy(fw, file); err != nil {
			endpoints.Logger.Warn(ctx, "Fehler beim Schreiben in ZIP-Datei: "+baseName)
			continue
		}
	}
}