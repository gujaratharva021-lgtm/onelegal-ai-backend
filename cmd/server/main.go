package main

import (
	"log"

	"legaltech-backend/internal/config"
	"legaltech-backend/internal/database"
	"legaltech-backend/internal/middleware"
	"legaltech-backend/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	database.Connect(cfg)

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	routes.Setup(router, cfg)

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
