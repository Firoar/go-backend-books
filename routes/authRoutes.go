package routes

import (
	"backend/Models"
	"backend/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"time"
)

// AllAuthRoutes registers all authentication-related routes.
func AllAuthRoutes(router *gin.RouterGroup, database *gorm.DB, jwtSecret []byte) {
	authRouter := router.Group("/auth")
	{
		authRouter.POST("/signup", func(c *gin.Context) {
			signUp(c, database)
		})
		authRouter.POST("/signin", func(c *gin.Context) {
			signIn(c, database, jwtSecret)
		})

	}
}

// signUp handles user registration
func signUp(c *gin.Context, database *gorm.DB) {
	var userInput Models.User
	if err := c.ShouldBindJSON(&userInput); err != nil { // Corrected ShouldBindBodyWithJSON to ShouldBindJSON
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var existingUser Models.User
	result := database.Where("email = ?", userInput.Email).First(&existingUser)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	userInput.Password = string(hashedPassword)

	if result.RowsAffected == 0 {
		// user does not exist
		newUser := Models.User{
			Email:    userInput.Email,
			Password: userInput.Password,
			Name:     userInput.Name,
			Address:  userInput.Address,
			Phone:    userInput.Phone,
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
}

// signIn handles user login
func signIn(c *gin.Context, database *gorm.DB, jwtSecret []byte) {
	var userInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&userInput); err != nil { // Corrected ShouldBindBodyWithJSON to ShouldBindJSON
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	var existingUser Models.User
	result := database.Where("email = ?", userInput.Email).First(&existingUser)

	if result.Error != nil || !utils.CheckPassword(existingUser.Password, userInput.Password) {
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
		"token": tokenString,
	})
}
