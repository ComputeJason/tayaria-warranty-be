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

	// Health check
	r.GET("/api/ping", handlers.Ping)

	// Public auth routes
	r.POST("/api/admin/login", handlers.AdminLogin)
	r.POST("/api/master/login", handlers.MasterLogin)

	// User routes (public, no auth middleware)
	userRoutes := r.Group("/api/user")
	{
		userRoutes.POST("/warranty", handlers.RegisterWarranty)
		userRoutes.GET("/warranties/car-plate/:carPlate", handlers.GetWarrantiesByCarPlate)
		userRoutes.GET("/warranties/valid/:carPlate", handlers.HasValidWarrantyByCarPlate)
		userRoutes.GET("/warranty/receipt/:id", handlers.GetWarrantyReceipt)
	}

	// Admin routes (protected)
	adminRoutes := r.Group("/api/admin")
	adminRoutes.Use(middleware.AdminMiddleware())
	{
		// Claim management (moved from user routes)
		adminRoutes.POST("/claim", handlers.CreateClaim)
		adminRoutes.GET("/claims", handlers.GetShopClaims)
		adminRoutes.POST("/claim/:id/close", handlers.CloseClaim)
	}

	// Master admin routes (protected)
	masterRoutes := r.Group("/api/master")
	masterRoutes.Use(middleware.MasterMiddleware())
	{
		// claim management
		masterRoutes.GET("/claims", handlers.GetAllClaims)
		masterRoutes.GET("/claim/:id", handlers.GetClaimInfoByID)
		masterRoutes.POST("/claim/:id/tag-warranty", handlers.TagWarrantyToClaim)
		masterRoutes.POST("/claim/:id/change-status", handlers.ChangeClaimStatus)
		masterRoutes.POST("/claim/:id/pending", handlers.ChangeClaimStatusToPending)
		// warranty management
		masterRoutes.GET("/warranties/valid/:carPlate", handlers.GetValidWarrantiesForTagging)
		// retail account management
		masterRoutes.POST("/account", handlers.CreateRetailAccount)
		masterRoutes.GET("/account", handlers.GetRetailAccounts)
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
