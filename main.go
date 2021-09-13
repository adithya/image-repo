package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

var GCPPkey []byte

// DB is the global connection pool for the database connection
var DB *gorm.DB

// Client is the global connection pool for the GCP SDK
var Client *storage.Client

// GCPProjectID is the project ID withing GCP, should be passed wherever a project ID is needed as an argument
const GCPProjectID = "shopify-challenge-image-repo"

func main() {
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("host=db user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), os.Getenv("PGADMIN_LISTEN_PORT")),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	// Load GCP private key
	GCPPkey, err = ioutil.ReadFile("gcp-private-key.pem")

	// Migrate the schema
	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&Photo{})

	// Connect to Google Cloud SDK
	ctx := context.Background()
	Client, err = storage.NewClient(ctx, option.WithCredentialsFile("gcp-service-acc-creds.json"))
	if err != nil {
		panic(err.Error())
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/GetVersion", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("0.1\n"))
	})

	userService := http.NewServeMux()
	userService.HandleFunc("/signup", Signup)
	userService.HandleFunc("/authenticate", Authenticate)
	userService.HandleFunc("/refresh", Refresh)
	userService.HandleFunc("/logout", Logout)
	mux.Handle("/user/", http.StripPrefix("/user", userService))

	photoService := http.NewServeMux()
	photoService.Handle("/upload", AuthenticateAndReturnUsername(http.HandlerFunc(Upload)))
	photoService.Handle("/edit/permissions", AuthenticateAndReturnUsername(http.HandlerFunc(ChangePermissions)))
	photoService.Handle("/delete", AuthenticateAndReturnUsername(http.HandlerFunc(Delete)))
	mux.Handle("/photo/", http.StripPrefix("/photo", photoService))

	feedService := http.NewServeMux()
	feedService.HandleFunc("/public", GetFeed)                                               // public photos from all users
	feedService.Handle("/home", AuthenticateAndReturnUsername(http.HandlerFunc(GetGallery))) // all photos uploaded by user (public + private)
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
