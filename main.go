package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"time"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

var GCPPkey []byte

// DB is the global connection pool for the database connection
var DB *gorm.DB

// Client is the global connection pool for the GCP SDK
var Client *storage.Client

var IsDebug bool

// GCPProjectID is the project ID withing GCP, should be passed wherever a project ID is needed as an argument
const GCPProjectID = "shopify-challenge-image-repo"

func main() {
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), os.Getenv("PGADMIN_LISTEN_PORT")),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	IsDebug, err = strconv.ParseBool(os.Getenv("IS_DEBUG"))
	if err != nil {
		panic("IS_DEBUG in .env must either be \"true\" or \"false\"")
	}

	// Load GCP private key
	GCPPkey = GetPrivateKeyFromGCPCredentialsFile("gcp-service-acc-creds.json")

	// Migrate the schema
	DB.AutoMigrate(&User{})
	DB.AutoMigrate(&Photo{})

	// Connect to Google Cloud SDK
	ctx := context.Background()
	if IsDebug {
		Client, err = storage.NewClient(context.TODO(), option.WithoutAuthentication(), option.WithEndpoint(fmt.Sprintf("http://%s:4443/storage/v1/", os.Getenv("CLOUD_STORAGE_HOST"))))
	} else {
		Client, err = storage.NewClient(ctx, option.WithCredentialsFile("gcp-service-acc-creds.json"))
	}
	if err != nil {
		panic(err)
	}

	// Make sure public bucket exists and create it if it doesn't
	// Only run in production as Google Cloud Storage emulator for local development does not support metadata retrieval
	// TODO: Setup terraform to handle settin up prod environment from scratch
	if !IsDebug {
		_, err = Client.Bucket(PUBLIC_BUCKET_NAME).Attrs(ctx)
		if err == storage.ErrBucketNotExist {
			bkt := Client.Bucket(PUBLIC_BUCKET_NAME)
			if err = bkt.Create(ctx, GCPProjectID, nil); err != nil {
				panic(err)
			}
		}
		if err != nil {
			panic(err)
		}
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
	photoService.Handle("/details", DetermineIfAuthenticated(http.HandlerFunc(GetPhotoDetails)))
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
