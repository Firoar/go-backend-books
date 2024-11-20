package Models

import (
	"github.com/lib/pq"
	"time"
)

type Book struct {
	ID        uint           `gorm:"primary_key;AUTO_INCREMENT"`
	Title     string         `gorm:"not null"`
	Author    string         `gorm:"not null"`
	Synopsis  string         `gorm:"not null"`
	Price     float64        `gorm:"not null"`
	SellerID  uint           `gorm:"not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ImageUrl  string         `gorm:"not null"`
	Category  pq.StringArray `gorm:"type:text[]"`
	CreatedAt time.Time      `gorm:"not null; default:now()"`
	User      User           `gorm:"foreignKey:SellerID;constraint:OnDelete:CASCADE;"`
}
