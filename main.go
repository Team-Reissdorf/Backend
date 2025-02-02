package main

import (
	"context"
	"github.com/LucaSchmitz2003/DatabaseFlow"
	"github.com/LucaSchmitz2003/FlowServer"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/LucaSchmitz2003/FlowWatch/otelHelper"
	"github.com/Team-Reissdorf/Backend/authHelper"
	"github.com/Team-Reissdorf/Backend/databaseUtils"
	"github.com/Team-Reissdorf/Backend/endpoints/athleteManagement"
	"github.com/Team-Reissdorf/Backend/endpoints/backendSettings"
	"github.com/Team-Reissdorf/Backend/endpoints/performanceManagement"
	"github.com/Team-Reissdorf/Backend/endpoints/ping"
	"github.com/Team-Reissdorf/Backend/endpoints/userManagement"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"os"
	"strconv"
)

var (
	tracer = otel.Tracer("MainTracer")
	logger = FlowWatch.GetLogHelper()

	frontendUrl string
)

func init() {
	ctx := context.Background()

	// Load the environment variables
	if err := godotenv.Load(".env"); err != nil {
		logger.Fatal(ctx, "Failed to load environment variables")
	}

	// Get the information if the program should run in production mode
	productionMode, err1 := strconv.ParseBool(os.Getenv("RELEASE_MODE"))
	if err1 != nil {
		err1 = errors.Wrap(err1, "Failed to parse RELEASE_MODE, using default")
		logger.Warn(ctx, err1)
		productionMode = false
	}
	if !productionMode {
		FlowWatch.SetLogLevel(FlowWatch.Debug)
		logger.Warn(ctx, "Development mode enabled. Please change before release!")
	}

	// Get the frontend url to allow as origin
	frontendUrl = os.Getenv("FRONTEND_URL")
	if frontendUrl == "" {
		err := errors.New("FRONTEND_URL is empty, using default")
		logger.Warn(ctx, err)
		frontendUrl = "http://localhost:8081"
	}

	// Register the models for the database
	DatabaseFlow.RegisterModels(ctx,
		databaseUtils.Trainer{},
		databaseUtils.Athlete{},
		databaseUtils.Discipline{},
		databaseUtils.Exercise{},
		databaseUtils.Performance{},
	)
	DatabaseFlow.GetDB(ctx) // Initialize the database connection

	// Initialize the OpenTelemetry SDK connection to the backend
	otelHelper.SetupOtelHelper()
}

func main() {
	ctx := context.Background()

	// Defer the shutdown function to ensure a graceful shutdown of the SDK connection at the end
	defer otelHelper.Shutdown()

	// Set frontend url as accepted origin for cors
	acceptedOrigins := []string{
		frontendUrl,
	}

	// Initialize the server
	address, router := FlowServer.InitServer(ctx, defineRoutes, acceptedOrigins)

	// Start the server and keep it alive
	keepAlive := FlowServer.StartServer(ctx, router, address)
	defer keepAlive()

	// ...
}

func defineRoutes(ctx context.Context, router *gin.Engine) {
	// Create a sub-span
	_, span := tracer.Start(ctx, "Define http routes")
	defer span.End()

	// Define the http routes
	v1 := router.Group("/v1") // Define a versioned route group
	{
		v1.GET("/ping", ping.Ping)
		v1.GET("/coffee", ping.Teapot)

		settings := v1.Group("/backendSettings", authHelper.GetAuthMiddlewareFor(authHelper.SettingsAccessToken))
		{
			settings.POST("/change-log-level", backendSettings.ChangeLogLevel) // ToDo: Add auth
		}

		user := v1.Group("/user")
		{
			user.POST("/register", userManagement.Register)
			user.POST("/login", userManagement.Login)
			user.POST("/start-session", authHelper.GetAuthMiddlewareFor(authHelper.RefreshToken), userManagement.StartSession)
			user.POST("/logout", userManagement.Logout)
		}

		athlete := v1.Group("/athlete", authHelper.GetAuthMiddlewareFor(authHelper.AccessToken))
		{
			athlete.POST("/create", athleteManagement.CreateAthlete)
			athlete.POST("/bulk-create", athleteManagement.CreateAthleteCVS)
			athlete.GET("/get-all", athleteManagement.GetAllAthletes)
			athlete.GET("/get-one/:AthleteId", athleteManagement.GetAthleteByID)
			athlete.PUT("/edit", athleteManagement.EditAthlete)
			athlete.DELETE("/delete/:AthleteId", athleteManagement.DeleteAthlete)
		}

		performance := v1.Group("/performance", authHelper.GetAuthMiddlewareFor(authHelper.AccessToken))
		{
			performance.POST("/create", performanceManagement.CreatePerformance)
		}
	}
}
