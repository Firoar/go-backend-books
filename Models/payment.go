package Models

// VerificationStatus represents the verification status of a user or company.
type VerificationStatus string
type DeliveredStatus string

const (
	NotDone VerificationStatus = "not done"
	Yes     VerificationStatus = "yes"
	No      VerificationStatus = "no"
)

const (
	Ok    DeliveredStatus = "ok"
	NotOk DeliveredStatus = "not ok"
)

// Payment represents a payment record in the database.
type Payment struct {
	PaymentID                uint               `gorm:"primary_key;AUTO_INCREMENT"` // Primary key for the payment
	UserID                   uint               `gorm:"not null"`                   // ID of the user who made the payment
	SellerID                 uint               `gorm:"not null"`                   // ID of the seller
	BookID                   uint               `gorm:"not null"`                   // ID of the book
	Price                    float64            `gorm:"not null"`                   // Price of the book
	TransactionID            string             `gorm:"not null;unique"`            // Unique transaction ID
	PhoneNumber              string             `gorm:"not null"`                   // Phone number of the payer
	CompanyVerifiedStatus    VerificationStatus `gorm:"default:'not done'"`         // Company verification status
	SellerVerificationStatus VerificationStatus `gorm:"default:'not done'"`         // Seller verification status
	SecretPassword           string             `gorm:"not null"`                   // Secret password generated during payment
	DeliveredStatus          DeliveredStatus    `gorm:"default:'not ok'"`
}
