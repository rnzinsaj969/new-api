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

	// Start server - bind address controlled by HOST env var.
	// Default to 127.0.0.1 for safety when running locally without Docker.
	// Set HOST=0.0.0.0 to expose on all interfaces (e.g. inside Docker or
	// when running behind a reverse proxy on another machine).
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	err = server.Run(host + ":" + port)
	if err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
