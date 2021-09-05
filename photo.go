package main

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"net/http"
)

// Photo represents a photo that has been uploaded
type Photo struct {
	// Each photo has an unique ID, that allows us to identify it in the users bucket
	ID string `json:"id" gorm:"primaryKey"`
	// Each photo can either be public or private, and is private by default
	IsPublic bool `json:"IsPublic" gorm:"default:false"`
	// Each photo is owned by a valid user from the users table
	UserID string `json:"UserId"`
	User   User
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
		IsPublic: false, // ******* TODO: DONT HARDCODE GET FROM FORM *****
		UserID:   *bucketID,
	}
	DB.Create(&photo)

	// Retrieve user's bucket
	bkt := Client.Bucket(*bucketID)

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

	w.WriteHeader(http.StatusOK)
}

// ChangePermissions allows users to change the visibility of photo between public (everyone can see) and private (only you can see)
func ChangePermissions(w http.ResponseWriter, r *http.Request) {

}

// Delete allows users to delete photos
func Delete(w http.ResponseWriter, r *http.Request) {

}
