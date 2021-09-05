package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func AuthenticateAndReturnUsername(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tknStr := c.Value

		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !tkn.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "username", claims.Username))
		next.ServeHTTP(w, r)
	})
}

func GetUserGUID(username string) (*string, error) {
	var users []User
	DB.Where(&User{Username: username}).Find(&users)

	if len(users) > 1 {
		return nil, fmt.Errorf("Multiple users returned")
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("No user found")
	}

	return &users[0].ID, nil
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
