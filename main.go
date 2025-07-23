package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tayaria-warranty-be/config"
	"tayaria-warranty-be/db"
	"tayaria-warranty-be/handlers"
	"tayaria-warranty-be/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// pingServer pings the server every 10 minutes to keep it active
func pingServer(port string) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// Determine the URL to ping
	var pingURL string
	if os.Getenv("RENDER") == "true" {
		// On Render, use the RENDER_EXTERNAL_URL environment variable
		renderURL := os.Getenv("RENDER_EXTERNAL_URL")
		if renderURL != "" {
			pingURL = renderURL + "/ping"
		} else {
			// Fallback to localhost if RENDER_EXTERNAL_URL is not set
			pingURL = "http://localhost:" + port + "/api/ping"
		}
	} else {
		// Local development
		pingURL = "http://localhost:" + port + "/api/ping"
	}

	log.Printf("🔗 Auto-ping will use URL: %s", pingURL)

	for {
		select {
		case <-ticker.C:
			resp, err := http.Get(pingURL)
			if err != nil {
				log.Printf("❌ Failed to ping server: %v", err)
			} else {
				resp.Body.Close()
				log.Printf("✅ Server pinged successfully at %s", time.Now().Format("15:04:05"))
			}
		}
	}
}

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
		userRoutes.GET("/warranties/car-plate/:carPlate", handlers.GetWarrantiesByCarPlate) // this is for User Warranty Check
		userRoutes.GET("/warranties/valid/:carPlate", handlers.HasValidWarrantyByCarPlate)
		userRoutes.GET("/warranty/receipt/:id", handlers.GetWarrantyReceipt)
		userRoutes.GET("/claim/:id/supporting-doc", handlers.GetWarrantySupportingDoc)
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
		masterRoutes.POST("/claim/:id/accept", handlers.ChangeClaimStatusToAccepted)
		masterRoutes.POST("/claim/:id/reject", handlers.ChangeClaimStatusToRejected)
		masterRoutes.POST("/claim/:id/supporting-doc", handlers.TagSupportingDocToClaim)
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

	// Start the ping goroutine to keep server active
	go pingServer(port)
	log.Printf("🚀 Auto-ping started - server will ping itself every 10 minutes to stay active on Render")

	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
