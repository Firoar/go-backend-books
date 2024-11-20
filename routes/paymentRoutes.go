package routes

import (
	"backend/Models"
	"backend/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func AllPaymentRoutes(router *gin.RouterGroup, database *gorm.DB) {
	paymentRouter := router.Group("/protected/payment")
	//paymentRouter.Use(Middlewares.AuthMiddlewares())
	{
		paymentRouter.GET("/test", func(c *gin.Context) {
			testPayment(c)
		})

		paymentRouter.POST("/do", func(c *gin.Context) {
			doPayment(c, database) // Pass database from the outer scope
		})

		paymentRouter.POST("/owner_approval/:id", func(c *gin.Context) {
			ownerApproval(c, database)
		})
		paymentRouter.POST("/seller_approval/:id", func(c *gin.Context) {
			sellerApproval(c, database)
		})
		paymentRouter.POST("/otp", func(c *gin.Context) {
			ApproveOtp(c, database)
		})

		paymentRouter.GET("/:id", func(c *gin.Context) {
			getPaymentDetails(c, database)
		})

		paymentRouter.GET("/get_payment_details/:id", func(c *gin.Context) {
			sendPaymentDetails(c, database)
		})

	}
}

func ApproveOtp(c *gin.Context, database *gorm.DB) {
	// Input struct for binding JSON data
	type UserInput struct {
		SellerEmail string `json:"sellerEmail"`
		BuyerEmail  string `json:"buyerEmail"`
		Otp         string `json:"otp"`
	}

	var userInput UserInput
	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	fmt.Printf("UserInput: %+v\n", userInput)

	var buyer Models.User
	var seller Models.User
	var payment Models.Payment
	var book Models.Book

	// Retrieve seller by email
	if result := database.Where("email = ?", userInput.SellerEmail).First(&seller); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "seller not found"})
		return
	}

	// Retrieve buyer by email
	if result := database.Where("email = ?", userInput.BuyerEmail).First(&buyer); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "buyer not found"})
		return
	}

	// Retrieve payment using buyer ID, OTP, and seller ID
	if result := database.Where("user_id = ? AND secret_password = ? AND seller_id = ?", buyer.ID, userInput.Otp, seller.ID).First(&payment); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "payment not found"})
		return
	}

	// Retrieve book using payment's BookID
	if result := database.Where("id = ?", payment.BookID).First(&book); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Book not found"})
		return
	}

	// Update payment status
	payment.DeliveredStatus = Models.Ok
	if err := database.Save(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Prepare email notifications
	ownerMailTo := "chiru02.dev@gmail.com"
	ownerSubject := "Product Delivered"

	// Construct the HTML email body for the owner
	ownerBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Product Delivered</title>
</head>
<body>
    <h2>Product Delivered</h2>
    <p><strong>%s</strong> sold by <strong>%s</strong> has been delivered to <strong>%s</strong>.</p>
    <p>Please do payment to: <strong>%s</strong></p>
    <p>Phone number: <strong>%s</strong></p>
</body>
</html>`,
		book.Title,
		seller.Email,
		buyer.Email,
		seller.Email,
		seller.Phone,
	)

	// Send email to owner
	if err := utils.SendMail(ownerMailTo, ownerSubject, ownerBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to owner"})
		return
	}

	// Prepare email notification for seller
	sellerMailTo := seller.Email
	sellerSubject := "Product Delivered"

	sellerMailBody := fmt.Sprintf(`
<html>
<body>
    <p>Dear %s,</p>
    <p>Your book '<strong>%s</strong>' was delivered to the buyer.</p>
    <p>You will receive the payment within 7 working days.</p>
</body>
</html>`,
		seller.Name, book.Title,
	)

	// Send email to seller
	if err := utils.SendMail(sellerMailTo, sellerSubject, sellerMailBody); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to seller"})
		return
	}

	// Respond with success message
	c.JSON(http.StatusOK, gin.H{"message": "Payment approved"})
}

func getPaymentDetails(c *gin.Context, database *gorm.DB) {
	useIdStr := c.Param("id")
	userId, err := strconv.ParseUint(useIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID format"})
		return
	}

	var payments []Models.Payment
	// Query to get all payments for the specified user ID
	if err := database.Where("user_id = ?", userId).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Error fetching payment details", "error": err.Error()})
		return
	}

	// Check if payments were found
	if len(payments) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No payment records found for this user"})
		return
	}

	// Return the payment details
	c.JSON(http.StatusOK, payments)
}

func sellerApproval(c *gin.Context, database *gorm.DB) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payment ID format"})
		return
	}
	approved := c.Query("approved") == "true"
	var payment Models.Payment
	if result := database.Where("payment_id = ?", paymentID).First(&payment); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Payment not found"})
		return
	}
	sellerId, buyerId, bookId := payment.SellerID, payment.UserID, payment.BookID

	var seller Models.User
	var buyer Models.User
	var book Models.Book

	if result := database.Where("id = ?", bookId).First(&book); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Book not found"})
		return
	}

	if result := database.Where("id = ?", sellerId).First(&seller); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Seller not found"})
		return
	}

	if result := database.Where("id = ?", buyerId).First(&buyer); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Buyer not found"})
		return
	}

	if approved {
		//Send mail to buyer with OTP
		buyerMailTo := buyer.Email
		buyerSubject := "Approved Succesfull"
		buyerBody := fmt.Sprintf("You will receive '%v' within 7 days.\n\rOTP = %v", book.Title, payment.SecretPassword)

		if err := utils.SendMail(buyerMailTo, buyerSubject, buyerBody); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to buyer"})
			return
		}

		// send mail to owner saying that pick up order
		ownerMailTo := "chiru02.dev@gmail.com"
		ownerSubject := fmt.Sprintf("Seller aprroved")

		ownerBody := fmt.Sprintf("the seller : %v has approved\nPick up from : %v\n\rDrop to : %v", seller.Email, seller.Address, buyer.Address)

		err = utils.SendMail(ownerMailTo, ownerSubject, ownerBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to owner"})
			return
		}
		payment.SellerVerificationStatus = Models.Yes
		if err := database.Save(&payment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("We will pick up the book : %v within 3 days at your address : %v", book.Title, seller.Address)})

	} else {
		buyerMailTo := buyer.Email
		buyerSubject := "Approved UnSuccesfull"
		buyerBody := fmt.Sprintf("You will receive refund within 7 days")

		if err := utils.SendMail(buyerMailTo, buyerSubject, buyerBody); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to buyer"})
			return
		}

		ownerMailTo := "chiru02.dev@gmail.com"
		ownerSubject := fmt.Sprintf("Seller un aprroved")

		ownerBody := fmt.Sprintf("refund to %v, phno : %v", buyer.Email, buyer.Phone)

		err = utils.SendMail(ownerMailTo, ownerSubject, ownerBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to owner"})
			return
		}

		payment.SellerVerificationStatus = Models.No
		if err := database.Save(&payment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Rejected successfully"})

	}

}

func ownerApproval(c *gin.Context, database *gorm.DB) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payment ID format"})
		return
	}
	approved := c.Query("approved") == "true"

	var payment Models.Payment
	if result := database.Where("payment_id = ?", paymentID).First(&payment); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Payment not found"})
		return
	}

	sellerId, buyerId, bookId := payment.SellerID, payment.UserID, payment.BookID

	var seller Models.User
	var buyer Models.User
	var book Models.Book

	if result := database.Where("id = ?", bookId).First(&book); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Book not found"})
		return
	}

	if result := database.Where("id = ?", sellerId).First(&seller); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Seller not found"})
		return
	}

	if result := database.Where("id = ?", buyerId).First(&buyer); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Buyer not found"})
		return
	}

	if approved {
		// Send email to seller
		sellerMailTo := seller.Email
		sellerSubject := "Approval for Picking Up the Order"
		sellerWebsiteLink := fmt.Sprintf("http://localhost:5000/approve-payment/%v", payment.PaymentID)

		sellerMailBody := fmt.Sprintf(
			`<html>
        <body>
            <p>Dear %v,</p>
            <p>The book '<strong>%v</strong>' which you put under sale has received a payment. Click the link below to send your approval:</p>
            <p><a href="%v">Approve Order</a></p>
        </body>
    </html>`,
			seller.Name, book.Title, sellerWebsiteLink,
		)

		if err := utils.SendMail(sellerMailTo, sellerSubject, sellerMailBody); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to seller"})
			return
		}

		// Send mail to buyer with OTP
		//buyerMailTo := buyer.Email
		//buyerSubject := ""
		//buyerBody := fmt.Sprintf("You will receive '%v' within 7 days.\nOTP = %v", book.Title, payment.SecretPassword)
		//
		//if err := utils.SendMail(buyerMailTo, buyerSubject, buyerBody); err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to buyer"})
		//	return
		//}
		payment.CompanyVerifiedStatus = Models.Yes
		if err := database.Save(&payment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		// Send confirmation message back
		c.JSON(http.StatusOK, gin.H{"message": "Payment approved successfully."})

	} else {
		// Don't send email to seller
		payment.CompanyVerifiedStatus = Models.No
		if err := database.Save(&payment).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		// Send email to buyer about the refund
		buyerMailTo := buyer.Email
		buyerSubject := "Refund initiated"
		buyerBody := "The transaction ID of the payment is invalid or you have done in sufficient payment. You will receive a refund from the bank within 7 working days."

		if err := utils.SendMail(buyerMailTo, buyerSubject, buyerBody); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to buyer"})
			return
		}

		// Send confirmation message back
		c.JSON(http.StatusOK, gin.H{"message": "Payment rejected successfully."})
	}
}

func sendPaymentDetails(c *gin.Context, database *gorm.DB) {
	payment := Models.Payment{}
	paymentIDStr := c.Param("id")

	// Parse the payment ID from the URL parameter
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid payment ID format",
		})
		return
	}

	// Query the database for the payment record
	result := database.Where("payment_id = ?", paymentID).First(&payment)
	if result.Error != nil {
		// Check if the error is due to no record found
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Payment not found",
			})
		} else {
			// Handle other database errors
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Internal server error",
			})
		}
		return
	}

	// Return the payment details if found
	c.JSON(http.StatusOK, gin.H{
		"payment": payment,
	})
}

func doPayment(c *gin.Context, database *gorm.DB) {
	// Payment logic goes here
	type userInput struct {
		UserId        uint    `json:"user_id"`
		SellerId      uint    `json:"seller_id"`
		BookId        uint    `json:"book_id"`
		Price         float64 `json:"price"`
		TransactionId string  `json:"transaction_id"`
		PhoneNumber   string  `json:"phone_number"`
	}

	var input userInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	fmt.Println("printing : ", input.SellerId, input.UserId, input.BookId)
	password, err := utils.GeneratePassword()
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error-pass"})

	}
	//hashPassword := utils.HashPassword(password)

	payment := &Models.Payment{
		UserID:         input.UserId,
		SellerID:       input.SellerId,
		BookID:         input.BookId,
		Price:          input.Price,
		TransactionID:  input.TransactionId,
		PhoneNumber:    input.PhoneNumber,
		SecretPassword: password,
	}

	if err := database.Create(&payment).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create payment"})
		return
	}

	var buyer Models.User
	if result := database.Where("id = ?", payment.UserID).First(&buyer); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "buyer not found"})
		return
	}
	var seller Models.User
	if result := database.Where("id = ?", payment.SellerID).First(&seller); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Seller not found"})
		return
	}

	var book Models.Book
	if result := database.Where("id = ?", payment.BookID).First(&book); result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "book not found"})
		return
	}

	//   now send a email to owner
	ownerMailTo := "chiru02.dev@gmail.com"
	ownerSubject := fmt.Sprintf("Approval for buying book  %v from %v", book.Title, buyer.Email)

	ownerBody := fmt.Sprintf(`
    <p>The transaction ID: %v</p>
    <p>Phone Number: %v</p>
    <p>Amount: %v</p>
    <p><a href="http://localhost:3001/approve-payment/%v">Approve it</a></p>
`, payment.TransactionID, payment.PhoneNumber, payment.Price, payment.PaymentID)

	err = utils.SendMail(ownerMailTo, ownerSubject, ownerBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email to owner"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": payment})

}

func testPayment(c *gin.Context) {
	c.JSON(200, gin.H{
		"msg": "good to go",
	})
}
