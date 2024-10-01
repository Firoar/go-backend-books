package main

import (
	"backend/Db"
	"backend/Models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	database := Db.Connect()
	database.AutoMigrate(&Models.User{})

	router := gin.Default()

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
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	})

	router.POST("/api/signin", func(c *gin.Context) {
		var userInput Models.User
		if err := c.ShouldBindBodyWithJSON(&userInput); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
		}

		var existingUser Models.User
		result := database.Where("username = ?", userInput.Username).First(&existingUser)

		if result.RowsAffected == 0 {
			// user does not exist
			newUser := Models.User{
				Username: userInput.Username,
				Password: userInput.Password,
			}
			database.Create(&newUser)
			c.JSON(201, gin.H{
				"message": "user created successfully",
			})
		} else {
			c.JSON(400, gin.H{
				"error": "User already exists",
			})
		}

	})

	router.Run(":8080")
}
