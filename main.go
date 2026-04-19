package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"new-api/common"
	"new-api/middleware"
	"new-api/model"
	"new-api/router"
)

func main() {
	// Load environment variables from .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Initialize common settings
	common.SetupLogger()
	common.SysLog("New API starting...")

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("Failed to initialize database: " + err.Error())
	}
	defer model.CloseDB()

	// Initialize Redis if configured
	err = common.InitRedisClient()
	if err != nil {
		common.SysLog("Redis not available, using memory cache: " + err.Error())
	}

	// Initialize options from database
	model.InitOptionMap()

	// Setup Gin engine
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	server.Use(middleware.CORS())

	// Register all routes
	router.SetRouter(server)

	// Determine port - prefer PORT env var, fallback to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	common.SysLog(fmt.Sprintf("Server listening on port %s", port))

	// Start server
	err = server.Run(":" + port)
	if err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
