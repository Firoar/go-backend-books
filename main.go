package main

import (
	"backend/Db"
	"backend/Middlewares"
	"backend/Models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

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
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	})

	router.POST("/api/signup", func(c *gin.Context) {
		var userInput Models.User
		if err := c.ShouldBindBodyWithJSON(&userInput); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		var existingUser Models.User
		result := database.Where("email = ?", userInput.Email).First(&existingUser)

		if result.RowsAffected == 0 {
			// user does not exist
			newUser := Models.User{
				Email:    userInput.Email,
				Password: userInput.Password,
				Name:     userInput.Name,
			}
			database.Create(&newUser)
			c.JSON(201, gin.H{
				"message": "user created successfully",
			})
			return
		} else {
			c.JSON(400, gin.H{
				"error": "User already exists",
			})
			return
		}

	})

	router.POST("/api/signin", func(c *gin.Context) {
		var userInput struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindBodyWithJSON(&userInput); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		var existingUser Models.User
		result := database.Where("email = ?", userInput.Email).First(&existingUser)

		if result.Error != nil || existingUser.Password != userInput.Password {
			c.JSON(400, gin.H{
				"message": "Invalid email or password",
			})
			return
		}

		claims := Models.Claims{
			Email: existingUser.Email,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "unable to generate jwt token",
			})
			return
		}

		c.SetCookie("token", tokenString, 3600*24, "/", "", true, true)
		c.JSON(200, gin.H{
			"message": tokenString,
		})
	})

	protected := router.Group("/api/protected")
	protected.Use(Middlewares.AuthMiddlewares())

	protected.GET("/hi", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hi, This is a protected route",
		})
	})
	router.Run(":8080")
}
