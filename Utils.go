package main

import (
	"cloud.google.com/go/storage"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type gcpCreds struct {
	ProjectType      []byte `json:"type"`
	ProjectID        string `json:"project_id"`
	PrivateKeyID     string `json:"private_key_id"`
	PrivateKey       string `json:"private_key"`
	ClientEmail      string `json:"client_email"`
	ClientID         string `json:"client_id"`
	AuthURI          string `json:"auth_uri"`
	TokenURI         string `json:"token_uri"`
	AuthProviderx509 string `json:"auth_provider_x509_cert_url"`
	Clientx509       string `json:"client_x509_cert_url"`
}

// GetPrivateKeyFromGCPCredentialsFile retrieves the private key from the Google Cloud Platform .json credentials file
func GetPrivateKeyFromGCPCredentialsFile(filename string) []byte {
	credsFile, err := os.Open(filename)
	if err != nil {
		panic(err.Error())
	}
	defer credsFile.Close()

	return parsePrivateKey(credsFile)
}

func parsePrivateKey(credsFile io.Reader) []byte {
	credsFileRaw, _ := ioutil.ReadAll(credsFile)

	var creds gcpCreds
	json.Unmarshal(credsFileRaw, &creds)

	if creds.PrivateKey == "" {
		panic("private key in Google Cloud Platform credentials file is empty")
	}

	return []byte(creds.PrivateKey)
}

// GetURLForImage retrieves the url for the image requested
// If running locally with IsDebug set to true, it will return a normal bucket URL as SignedURLs are difficult to make work with Google Cloud Storage Emulator
// If running in production with IsDebug set to false, SignedURLs will be returned with a 5 hour expiry
func GetURLForImage(photo Photo) (string, error) {
	if IsDebug {
		return getBucketURLForImage(photo)
	}

	return getSignedURLForImage(photo)
}

func getBucketURLForImage(photo Photo) (string, error) {
	return "http://localhost:4443/" + getBucketForPhoto(photo) + "/" + photo.ID, nil
}

func getSignedURLForImage(photo Photo) (string, error) {
	return storage.SignedURL(getBucketForPhoto(photo), photo.ID, &storage.SignedURLOptions{
		GoogleAccessID: "cloud-storage-user@shopify-challenge-image-repo.iam.gserviceaccount.com",
		PrivateKey:     GCPPkey,
		Method:         "GET",
		Expires:        time.Now().Add(5 * time.Hour),
	})
}
