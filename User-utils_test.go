package main

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	password2 := "password"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	hashedPassword2, err := HashPassword(password2)
	if err != nil {
		t.Fatal(err)
	}

	if hashedPassword == hashedPassword2 {
		t.Fatalf("Hash is coliding easily")
	}

	if password == hashedPassword {
		t.Fatalf("Not hashing password properly")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	IsValid := CheckPasswordHash(password, hashedPassword)
	if !IsValid {
		t.Fatalf("brcypt not finding hash match when there is one")
	}

	IsValid = CheckPasswordHash("notpassword", hashedPassword)
	if IsValid {
		t.Fatalf("bcrypt finding hash match when there isn't one")
	}
}
