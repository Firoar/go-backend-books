package Models

type User struct {
	ID       uint   `gorm:"primary_key;AUTO_INCREMENT"`
	Password string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
	Name     string `gorm:"not null"`
	Books    []Book `gorm:"foreignKey:SellerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Address  string `gorm:"not null"`
	Phone    string `gorm:"not null"`
}
