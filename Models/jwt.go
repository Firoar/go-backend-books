package Models

import "github.com/dgrijalva/jwt-go"

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}
