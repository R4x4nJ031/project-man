package main

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func main() {

	secret := []byte("supersecretkey") // must match JWT_SECRET in docker-compose

	claims := jwt.MapClaims{
		"user_id": "123",
		"email":   "test@example.com",
		"iss":     "project-man",
		"aud":     "project-man-users",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		panic(err)
	}

	fmt.Println(signedToken)
}
