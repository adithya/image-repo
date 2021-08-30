package main

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))
var db *gorm.DB

type User struct {
	ID       uint   `gorm:"primaryKey"` // make sure this gets generated automatically
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

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

func SearchForExistingUser(username string, db *gorm.DB) (*[]User, error) {
	var users []User
	db.Where(&User{Username: username}).Find(&users)

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

func Welcome(w http.ResponseWriter, r *http.Request) {
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

	w.Write([]byte(fmt.Sprintf("Welcome %s!", claims.Username)))
}

func main() {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("host=db user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), os.Getenv("PGADMIN_LISTEN_PORT")),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&User{})

	mux := http.NewServeMux()
	mux.HandleFunc("/GetVersion", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("0.1\n"))
	})

	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		u, err := ValidateAndDecodeRequestBody(r)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(400)
			// http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		users, err := SearchForExistingUser(u.Username, db)
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

		db.Create(&u)

		w.Write([]byte("User created"))
		w.WriteHeader(201)
	})

	mux.HandleFunc("/authenticate", func(w http.ResponseWriter, r *http.Request) {
		u, err := ValidateAndDecodeRequestBody(r)
		if err != nil {
			w.Write([]byte(err.Error()))
			w.WriteHeader(400)
			// http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		users, err := SearchForExistingUser(u.Username, db)
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
			Value:    tokenString,
			HttpOnly: true,
			Expires:  expirationTime,
		})
		w.Write([]byte("Authentication successful"))
		w.WriteHeader(200)
	})

	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
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
			Value:    tokenString,
			HttpOnly: true,
			Expires:  expirationTime,
		})
	})

	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
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
			Value:    "",
			HttpOnly: true,
			Expires:  expirationTime,
		})
	})

	mux.HandleFunc("/welcome", Welcome)
	// Upload photo
	// Change photo privacy
	// delete photo
	// get public gallery
	// get complete gallery (including private photos)

	s := http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}

	fmt.Printf("Starting server at :8080")
	err = s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
