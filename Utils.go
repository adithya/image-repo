package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
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

	credsFileRaw, _ := ioutil.ReadAll(credsFile)

	var creds gcpCreds
	json.Unmarshal(credsFileRaw, &creds)

	if creds.PrivateKey == "" {
		panic("private key in Google Cloud Platform credentials file is empty")
	}

	return []byte(creds.PrivateKey)
}
