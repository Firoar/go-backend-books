package Models

type User struct {
	ID       uint   `gorm:"primary_key;AUTO_INCREMENT"`
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
}
