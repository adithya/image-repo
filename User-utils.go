package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateAndDecodeRequestBody(r *http.Request) (*User, error) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("Missing username or password")
	}

	if user.Username == "" {
		return nil, fmt.Errorf("Username not provided in request body")
	}

	if user.Password == "" {
		return nil, fmt.Errorf("Password not provided in request body")
	}

	return &user, nil
}

func SearchForExistingUser(username string) (*[]User, error) {
	var users []User
	DB.Where(&User{Username: username}).Find(&users)

	if len(users) > 1 {
		return nil, fmt.Errorf("Multiple users returned")
	}

	return &users, nil
}

func UserExists(users *[]User) bool {
	if len(*users) > 0 {
		return true
	}

	return false
}
