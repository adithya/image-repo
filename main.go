package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))
var DB *gorm.DB

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
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("host=db user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), os.Getenv("PGADMIN_LISTEN_PORT")),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&Photo{})

	mux := http.NewServeMux()
	mux.HandleFunc("/GetVersion", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("0.1\n"))
	})

	mux.HandleFunc("/welcome", Welcome)

	userService := http.NewServeMux()
	userService.HandleFunc("/signup", Signup)
	userService.HandleFunc("/authenticate", Authenticate)
	userService.HandleFunc("/refresh", Refresh)
	userService.HandleFunc("/logout", Logout)
	mux.Handle("/user/", http.StripPrefix("/user", userService))

	photoService := http.NewServeMux()
	photoService.HandleFunc("/upload", Upload)
	photoService.HandleFunc("/edit/permissions", ChangePermissions)
	photoService.HandleFunc("/delete", Delete)
	mux.Handle("/photo/", http.StripPrefix("/photo", photoService))

	feedService := http.NewServeMux()
	feedService.HandleFunc("/public", GetFeed)  // public photos from all users
	feedService.HandleFunc("/home", GetGallery) // all photos uploaded by user (public + private)
	mux.Handle("/feed/", http.StripPrefix("/feed", feedService))

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
