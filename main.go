package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"tayaria-warranty-be/config"
	"tayaria-warranty-be/db"
	"tayaria-warranty-be/handlers"
	"tayaria-warranty-be/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	if err := config.Init(); err != nil {
		log.Fatal("Failed to initialize config:", err)
	}

	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Set up signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server, closing DB pool...")
		db.Close()
		os.Exit(0)
	}()

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Test routes
	r.GET("/api/test/user/:uuid", handlers.GetUserByUUID)
	r.GET("/api/test/user/name/:name", handlers.GetUserByName)

	// Health check
	r.GET("/api/ping", handlers.Ping)

	// Public routes
	r.POST("/api/auth/login", handlers.LoginUser)
	r.POST("/api/auth/logout", handlers.LogoutUser)
	// Admin login should be public
	r.POST("/api/admin/login", handlers.AdminLogin)

	// User routes
	userRoutes := r.Group("/api/user")
	userRoutes.Use(middleware.AuthMiddleware())
	{
		userRoutes.GET("/:phoneNumber", handlers.GetUserInformation)
		userRoutes.PUT("/:phoneNumber", handlers.EditUserInformation)
		userRoutes.POST("/warranty", handlers.RegisterWarranty)
		userRoutes.GET("/warranties/car-plate/:carPlate", handlers.GetWarrantiesByCarPlate)
		userRoutes.GET("/warranties/valid/:carPlate", handlers.GetValidWarrantyByCarPlate)
		userRoutes.GET("/warranty/receipt/:id", handlers.GetWarrantyReceipt)
		userRoutes.POST("/claim", handlers.CreateClaim)
		userRoutes.GET("/claims", handlers.GetClaims)
		userRoutes.POST("/claim/tag-warranty", handlers.TagWarrantyToClaim)
		userRoutes.POST("/claim/change-status", handlers.ChangeClaimStatus)
		userRoutes.POST("/claim/close", handlers.CloseClaim)
	}

	// Admin routes
	adminRoutes := r.Group("/api/admin")
	adminRoutes.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
	{
		// TODO: Add admin-protected routes here as they are implemented
	}

	// Get port from environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
