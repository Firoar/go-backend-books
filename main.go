package main

import (
	"backend/Db"
	"backend/Models"
	"backend/routes"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// migrating models
func main() {
	database := Db.Connect()
	if err := database.AutoMigrate(&Models.User{}); err != nil {
		panic(fmt.Sprintf("auto migrate failed for user: %v", err))
	}
	if err := database.AutoMigrate(&Models.Book{}); err != nil {
		panic(fmt.Sprintf("auto migrate failed for book: %v", err))
	}
	if err := database.AutoMigrate(&Models.Payment{}); err != nil {
		panic(fmt.Sprintf("auto migrate failed for user: %v", err))
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3001", "http://localhost:5000", "http://localhost:8001"}, // Allowed origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},                                                          // Allowed methods
		AllowHeaders:     []string{"Authorization", "Content-Type"},                                                                    // Allowed headers
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Om nama shivaya",
			"data":    "Jai Sri Ram",
		})
	})

	router.GET("/api/allusers", func(c *gin.Context) {
		var users []Models.User
		results := database.Find(&users)

		if results.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to fetch all users",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	})

	apiRouter := router.Group("/api")
	routes.AllAuthRoutes(apiRouter, database, jwtSecret)
	routes.AllProtectedRoutes(apiRouter, database, jwtSecret)
	routes.AllPaymentRoutes(apiRouter, database)

	if err := router.Run(":8080"); err != nil {
		panic(fmt.Sprintf("Error starting server: %v", err))
	}
}
