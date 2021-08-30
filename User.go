package main

import (
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"time"
)

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type User struct {
	ID       uint   `gorm:"primaryKey"` // make sure this gets generated automatically
	Username string `json:"username"`
	Password string `json:"password"`
}

func Signup(w http.ResponseWriter, r *http.Request) {
	u, err := ValidateAndDecodeRequestBody(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(400)
		// http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users, err := SearchForExistingUser(u.Username)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(500)
	}

	if UserExists(users) {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		w.Write([]byte("Username already taken"))
		w.WriteHeader(400)
		return
	}

	hash, _ := HashPassword(u.Password)
	u.Password = hash

	DB.Create(&u)

	w.Write([]byte("User created"))
	w.WriteHeader(201)
}

func Authenticate(w http.ResponseWriter, r *http.Request) {
	u, err := ValidateAndDecodeRequestBody(r)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(400)
		// http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	users, err := SearchForExistingUser(u.Username)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(500)
	}

	if !UserExists(users) {
		w.Write([]byte("Incorrect username or password"))
		w.WriteHeader(401)
		// http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	passwordExists := CheckPasswordHash(u.Password, (*users)[0].Password)

	if !passwordExists {
		w.Write([]byte("Incorrect username or password"))
		w.WriteHeader(401)
		// http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: u.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Path:     "/",
		Value:    tokenString,
		HttpOnly: true,
		Expires:  expirationTime,
	})
	w.Write([]byte("Authentication successful"))
	w.WriteHeader(200)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
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

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Path:     "/",
		Value:    tokenString,
		HttpOnly: true,
		Expires:  expirationTime,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// logout logic
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

	expirationTime := time.Unix(0, 0)
	claims.ExpiresAt = expirationTime.Unix()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Path:     "/",
		Value:    "",
		HttpOnly: true,
		Expires:  expirationTime,
	})
}
