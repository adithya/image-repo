package main

import (
	"net/http"
)

// Photo represents a photo that has been uploaded
type Photo struct {
	// Each photo has an unique ID, that allows us to identify it in the users bucket
	ID uint `json:"id" gorm:"primaryKey"`
	// Each photo can either be public or private, and is private by default
	IsPublic bool `json:"IsPublic" gorm:"default:false"`
	// Each photo is owned by a valid user from the users table
	UserID uint `json:"UserId"`
	User   User
}

// Upload allows users to upload photos, they may be marked as public or private
func Upload(w http.ResponseWriter, r *http.Request) {

}

// ChangePermissions allows users to change the visibility of photo between public (everyone can see) and private (only you can see)
func ChangePermissions(w http.ResponseWriter, r *http.Request) {

}

// Delete allows users to delete photos
func Delete(w http.ResponseWriter, r *http.Request) {

}
