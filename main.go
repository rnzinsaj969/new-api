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
	// Using 3000 as default; 8080 and 8081 are often taken on my machine
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	common.SysLog(fmt.Sprintf("Server listening on port %s", port))

	// Start server - always bind to all interfaces so Docker/remote access works
	// Previously defaulted to 127.0.0.1 for local dev, but this caused issues
	// when running inside containers. Override with HOST env var if needed.
	// Defaulting to localhost for local dev; set HOST=0.0.0.0 when deploying.
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	err = server.Run(host + ":" + port)
	if err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
