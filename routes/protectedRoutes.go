package routes

import (
	"backend/Middlewares"
	"backend/Models"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

func AllProtectedRoutes(router *gin.RouterGroup, database *gorm.DB, jwtSecret []byte) {
	// Create the protected routes group
	protectedRoutes := router.Group("/protected")
	protectedRoutes.Use(Middlewares.AuthMiddlewares())
	{
		protectedRoutes.GET("/hi", func(c *gin.Context) {
			test(c)
		})

		// Move this route outside the protectedRoutes group
		protectedRoutes.GET("/book/:id", func(c *gin.Context) {
			getSingleBook(c, database)
		})

		protectedRoutes.POST("/book", func(c *gin.Context) {
			postSingleBook(c, database)
		})

		protectedRoutes.GET("/sellersbooks", func(c *gin.Context) {
			sellersBooks(c, database)
		})

		protectedRoutes.DELETE("/sellersbooks/:id", func(c *gin.Context) {
			deleteSellerBook(c, database)
		})

		protectedRoutes.PUT("/userprofilechange/", func(c *gin.Context) {
			changeProfileData(c, database)
		})

		protectedRoutes.GET("/userprofileInfo", func(c *gin.Context) {
			getMeUserInf0(c, database)
		})
	}

	// Allow unauthenticated access to /allbooks
	router.GET("/protected/allbooks", func(c *gin.Context) {
		getAllBooks(c, database)
	})

}

func deleteSellerBook(c *gin.Context, database *gorm.DB) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "enter the correct book id"})
		return
	}

	if err := database.Delete(&Models.Book{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete the book"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Book deleted successfully"})

}

func getMeUserInf0(c *gin.Context, database *gorm.DB) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	var existingUser Models.User
	if err := database.First(&existingUser, "email = ?", email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := gin.H{
		"user": gin.H{
			"email":   existingUser.Email,
			"address": existingUser.Address,
			"id":      existingUser.ID,
		},
	}

	c.JSON(http.StatusOK, response)

}

func changeProfileData(c *gin.Context, database *gorm.DB) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var userInput struct {
		Address  string `json:"address"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})

	}

	// now change the address of the user
	var existingUser Models.User
	if err := database.First(&existingUser, "email = ?", email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	existingUser.Address = userInput.Address
	userInput.Password = strings.TrimSpace(userInput.Password)
	if userInput.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		userInput.Password = string(hashedPassword)
		existingUser.Password = userInput.Password
	}
	if err := database.Save(&existingUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
		return
	}
	c.JSON(200, gin.H{
		"message": "User changed profile data",
	})

}

func sellersBooks(c *gin.Context, database *gorm.DB) {
	// Retrieve email from context
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email not found in context"})
		return
	}

	var user Models.User // Assuming you have a User model
	// Fetch user ID based on the email
	if err := database.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Now that we have the user ID, fetch books for this user
	var books []Models.Book
	if err := database.Where("seller_id = ?", user.ID).Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Something went wrong, failed to fetch books",
		})
		return
	}

	// Return the books in the response
	c.JSON(http.StatusOK, gin.H{
		"books": books,
	})
}

func postSingleBook(c *gin.Context, database *gorm.DB) {
	var userInput struct {
		Title    string   `json:"title"`
		Author   string   `json:"author"`
		Synopsis string   `json:"synopsis"`
		Price    float64  `json:"price"`
		ImageUrl string   `json:"image_url"`
		Category []string `json:"category"`
	}

	// Bind JSON input to userInput struct
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Fetch user by email from context
	var user Models.User
	userEmail, _ := c.Get("email")
	if err := database.Where("email = ?", userEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if the book already exists for the user
	for _, book := range user.Books {
		if book.Title == userInput.Title {
			c.JSON(http.StatusConflict, gin.H{"message": "Book already exists"})
			return
		}
	}

	// Create new book
	newBook := Models.Book{
		Title:    strings.ToUpper(userInput.Title),
		Author:   userInput.Author,
		Synopsis: userInput.Synopsis,
		Price:    userInput.Price,
		Category: userInput.Category,
		ImageUrl: userInput.ImageUrl,
		SellerID: user.ID, // Associate with the seller (user)
	}

	// Save the new book to the database
	if err := database.Create(&newBook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create book", "error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Book created successfully",
		"book_id": newBook.ID,
		// Optional: include the updated list of books if needed
	})
}

func getSingleBook(c *gin.Context, database *gorm.DB) {
	idStr := c.Param("id")
	// check typ of id , if not of type unit error
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var book Models.Book
	err = database.Preload("User").First(&book, id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("book with id %v not found", id)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"book": book,
	})
}

func test(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Jai Sri Ram, from protected route",
	})
}

func getAllBooks(c *gin.Context, database *gorm.DB) {
	var books []Models.Book
	result := database.Order("created_at desc").Find(&books)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	var sendBackData []struct {
		ID          uint     `json:"id"`
		Title       string   `json:"title"`
		Author      string   `json:"author"`
		Description string   `json:"description"`
		SellerID    uint     `json:"sellerId"`
		Price       float64  `json:"price"`
		ImageUrl    string   `json:"imageUrl"`
		Category    []string `json:"category"`
	}

	for _, book := range books {
		data := struct {
			ID          uint     `json:"id"`
			Title       string   `json:"title"`
			Author      string   `json:"author"`
			Description string   `json:"description"`
			SellerID    uint     `json:"sellerId"`
			Price       float64  `json:"price"`
			ImageUrl    string   `json:"imageUrl"`
			Category    []string `json:"category"`
		}{
			ID:          book.ID,
			Title:       book.Title,
			Author:      book.Author,
			Description: book.Synopsis,
			SellerID:    book.SellerID,
			Price:       book.Price,
			ImageUrl:    book.ImageUrl,
			Category:    book.Category,
		}
		sendBackData = append(sendBackData, data)
	}

	// Return the structured data in the response
	c.JSON(http.StatusOK, gin.H{
		"data": sendBackData,
	})
}
