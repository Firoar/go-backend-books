package Models

type User struct {
	ID       uint   `gorm:"primary_key;AUTO_INCREMENT"`
	Password string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
	Name     string `gorm:"not null"`
}
