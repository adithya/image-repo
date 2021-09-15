package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"time"
)

const PUBLIC_BUCKET_NAME = "shopify-image-repo_public"

// Photo represents a photo that has been uploaded
type Photo struct {
	// Each photo has an unique ID
	ID string `json:"PhotoID" gorm:"primaryKey"`
	// Each photo can either be public or private, and is private by default
	IsPublic bool `json:"IsPublic" gorm:"default:false"`
	// Each photo is owned by a valid user from the users table
	UserID string `json:"-"`
	User   User   `json:"-"`
	// For client side use
	ImageURL         string `json:"ImageURL" gorm:"-"`
	Username         string `json:"Username" gorm:"-"`
	IsOwnedByAPIUser bool   `json:"IsOwnedByAPIUser" gorm:"-"`
}

// GetPhotoDetails returns
func GetPhotoDetails(w http.ResponseWriter, r *http.Request) {
	// Determine the photo the user is requesting details of
	var requestedPhoto Photo
	err := json.NewDecoder(r.Body).Decode(&requestedPhoto)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Malformded json"))
		return
	}

	// Retrieve photo
	var photo Photo
	DB.Where(&Photo{ID: requestedPhoto.ID}).First(&photo)
	if photo.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("photo with id not found"))
		return
	}

	// Determine if the user is authenticated
	IsAuthenticated := r.Context().Value("IsAuthenticated").(bool)

	// If authenticated get username/id
	var username string
	var userID *string
	IsOwnedByAPIUser := false // Determine if user requesting image owns photo
	if IsAuthenticated {
		username = r.Context().Value("username").(string)
		userID, err = GetUserGUID(username)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(""))
		}

		if photo.UserID == *userID {
			IsOwnedByAPIUser = true
		}
	}

	// Fill in some values for use by client side
	photo.Username = GetUsernameForUser(photo.UserID)
	photo.IsOwnedByAPIUser = IsOwnedByAPIUser

	if photo.IsPublic /* can return image regardless of who is requesting */ {
		url, err := storage.SignedURL(PUBLIC_BUCKET_NAME, photo.ID, &storage.SignedURLOptions{
			GoogleAccessID: "cloud-storage-user@shopify-challenge-image-repo.iam.gserviceaccount.com",
			PrivateKey:     GCPPkey,
			Method:         "GET",
			Expires:        time.Now().Add(5 * time.Hour),
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		photo.ImageURL = url

		photoItem, err := json.Marshal(photo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(photoItem)
	} else /* need to verify owner is requesting image */ {
		if IsOwnedByAPIUser {
			url, err := storage.SignedURL(getBucketForPhoto(photo), photo.ID, &storage.SignedURLOptions{
				GoogleAccessID: "cloud-storage-user@shopify-challenge-image-repo.iam.gserviceaccount.com",
				PrivateKey:     GCPPkey,
				Method:         "GET",
				Expires:        time.Now().Add(5 * time.Hour),
			})

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}

			photo.ImageURL = url

			photoItem, err := json.Marshal(photo)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(photoItem)
		}
	}

	w.WriteHeader(http.StatusBadRequest)
}

// Upload allows users to upload photos, they may be marked as public or private
func Upload(w http.ResponseWriter, r *http.Request) {
	// Get uploaded file
	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("uploadFile")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Get isPublic attribute
	IsPublicFromValue := r.FormValue("IsPublic")
	if IsPublicFromValue == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	IsPublic, err := strconv.ParseBool(IsPublicFromValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Identify who the user is
	username := r.Context().Value("username")
	if username == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user bucket id
	bucketID, err := GetUserGUID(username.(string))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate a unique ID to identify the photo object
	photoID := uuid.New().String()

	// Register photo in photos table
	photo := Photo{
		ID:       photoID,
		IsPublic: IsPublic,
		UserID:   *bucketID,
	}
	DB.Create(&photo)

	// Retrieve user's bucket
	bkt := Client.Bucket(getBucketForPhoto(photo))

	// Upload photo to bucket
	obj := bkt.Object(photoID)
	objWriter := obj.NewWriter(r.Context())
	if _, err := io.Copy(objWriter, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := objWriter.Close(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(photoID))
	w.WriteHeader(http.StatusOK)
}

// ChangePermissions allows users to change the visibility of photo between public (everyone can see) and private (only you can see)
func ChangePermissions(w http.ResponseWriter, r *http.Request) {
	// Identify who the user is
	username := r.Context().Value("username")
	if username == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get userid for user
	userID, err := GetUserGUID(username.(string))
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Retrieve PhotoID and IsPublic from JSON request body
	var requestedPhoto Photo
	err = json.NewDecoder(r.Body).Decode(&requestedPhoto)
	if err != nil {
		w.Write([]byte("Missing PhotoID or IsPublic attribute"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if requestedPhoto.ID == "" {
		w.Write([]byte("PhotoID not provided in request body"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// make sure photo exists
	var photos []Photo
	DB.Where(&Photo{ID: requestedPhoto.ID}).Find(&photos)

	if len(photos) > 1 {
		w.Write([]byte("Multiple photos returned"))
		w.WriteHeader(http.StatusInternalServerError)

	}

	if len(photos) == 0 {
		w.Write([]byte("No photos returned"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	photo := photos[0]

	// Make sure photo belongs to user
	if photo.UserID != *userID {
		w.Write([]byte("photo does not belong to user"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// If permission has changed photo needs to be updated in photos tabe and object needs to be moved between buckets
	if photo.IsPublic != requestedPhoto.IsPublic {
		// If permission has gone from public to private
		if photo.IsPublic == true && requestedPhoto.IsPublic == false {
			err = moveBuckets(r.Context(), PUBLIC_BUCKET_NAME, *userID, photo.ID)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// If permission has gone from private to public
		if photo.IsPublic == false && requestedPhoto.IsPublic == true {
			err = moveBuckets(r.Context(), *userID, PUBLIC_BUCKET_NAME, photo.ID)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// change permission for photo in photos table
		photo.IsPublic = requestedPhoto.IsPublic
		DB.Save(&photo)
	}

	w.Write([]byte("photo visibility has been changed"))
	w.WriteHeader(http.StatusOK)
	return
}

func moveBuckets(ctx context.Context, srcBucketName string, dstBucketName string, objName string) error {
	src := Client.Bucket(srcBucketName).Object(objName)
	dst := Client.Bucket(dstBucketName).Object(objName)

	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("Object(%q).CopierFrom(%q).Run: %v", objName, srcBucketName, err)
	}

	if err := src.Delete(ctx); err != nil {
		return fmt.Errorf("Object(%q).Delete: %v", objName, err)
	}

	return nil
}

// Delete allows users to delete photos
func Delete(w http.ResponseWriter, r *http.Request) {
	// get user info
	username := r.Context().Value("username")
	if username == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// retrieve photo id from api call
	var requestedPhoto Photo
	err := json.NewDecoder(r.Body).Decode(&requestedPhoto)
	if err != nil {
		w.Write([]byte("Missing PhotoID or IsPublic attribute"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if requestedPhoto.ID == "" {
		w.Write([]byte("PhotoID not provided in request body"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// make sure photo exists
	var photos []Photo
	DB.Where(&Photo{ID: requestedPhoto.ID}).Find(&photos)

	if len(photos) > 1 {
		w.Write([]byte("Multiple photos returned"))
		w.WriteHeader(http.StatusInternalServerError)

	}

	if len(photos) == 0 {
		w.Write([]byte("No photos returned"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	photo := photos[0]

	// Make sure photo belongs to user
	userID, err := GetUserGUID(username.(string))
	if photo.UserID != *userID {
		w.Write([]byte("photo does not belong to user"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// delete photo from photos table
	DB.Delete(&photo)

	// delete file from bucket
	imageFile := Client.Bucket(getBucketForPhoto(photo)).Object(photo.ID)
	if err = imageFile.Delete(r.Context()); err != nil {
		err = fmt.Errorf("Object(%q).Delete: %v", photo.ID, err)
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("photo deleted"))
	w.WriteHeader(http.StatusOK)
}

func getBucketForPhoto(photo Photo) string {
	if photo.IsPublic {
		return PUBLIC_BUCKET_NAME
	}

	return photo.UserID
}
