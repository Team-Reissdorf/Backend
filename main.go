package main

import (
	"context"
	"github.com/LucaSchmitz2003/FlowServer"
	"github.com/LucaSchmitz2003/FlowWatch"
	"github.com/LucaSchmitz2003/FlowWatch/otelHelper"
	"github.com/Team-Reissdorf/Backend/endpoints"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("MainTracer")
	logger = FlowWatch.GetLogHelper()
)

func init() {
	ctx := context.Background()

	// Load the environment variables
	if err := godotenv.Load(".env"); err != nil {
		logger.Fatal(ctx, "Failed to load environment variables")
	}

	// Register the models for the database
	/*databaseHelper.RegisterModels(
		ctx,
	)
	databaseHelper.GetDB(ctx) // Initialize the database connection*/

	// Initialize the OpenTelemetry SDK connection to the backend
	otelHelper.SetupOtelHelper()
}

func main() {
	ctx := context.Background()

	// Defer the shutdown function to ensure a graceful shutdown of the SDK connection at the end
	defer otelHelper.Shutdown()

	// Initialize the server
	address, router := FlowServer.InitServer(ctx, defineRoutes)

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
		v1.GET("/ping", endpoints.Ping)
		v1.GET("/coffee", endpoints.Teapot)
	}
}
