package main

import (
	"cloud.google.com/go/storage"
	"encoding/json"
	"net/http"
	"time"
)

// FeedItem regardless of source
type FeedItem struct {
	// Each photo has an unique ID, that allows us to identify it in the users bucket
	ID       string `json:"PhotoID"`
	ImageURL string `json:"ImageURL"`
}

// GetFeed returns all photos that have public permissions
// TODO: Pagination of results
func GetFeed(w http.ResponseWriter, r *http.Request) {
	// Get array of photos with isPublic set to true
	var photos []Photo
	result := DB.Where(&Photo{IsPublic: true}).Find(&photos)
	if result.Error != nil {
		w.Write([]byte(result.Error.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var items []FeedItem
	// Loop through photos array
	for _, photo := range photos {
		// Get url for each object in photos array
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

		// Insert photo ID and image url into feed item
		items = append(items, FeedItem{ID: photo.ID, ImageURL: url})
	}

	// Return array of photo structs via json
	feed, err := json.Marshal(items)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	// 200 OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(feed)
}

// GetGallery return all public and private photos owned by the user
func GetGallery(w http.ResponseWriter, r *http.Request) {
	// Identify who the user is
	username := r.Context().Value("username")
	if username == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get user id for query
	userID, err := GetUserGUID(username.(string))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get array of photos owned by user (public or private)
	var photos []Photo
	result := DB.Where(&Photo{UserID: *userID}).Find(&photos)
	if result.Error != nil {
		w.Write([]byte(result.Error.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var items []FeedItem
	// Loop through photos array
	for _, photo := range photos {
		// Get url for each object in photos array
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

		// Insert photo ID and image url into feed item
		items = append(items, FeedItem{ID: photo.ID, ImageURL: url})
	}

	// Return array of photo structs via json
	feed, err := json.Marshal(items)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	// 200 OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(feed)
}
