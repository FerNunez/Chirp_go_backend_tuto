package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashingPassword(t *testing.T) {

	inputPass := "FerFer123"

	hash, err := HashPassword(inputPass)
	if err != nil {
		t.Fatalf("Could not hash input password. Err: %v", err)
		return
	}

	if CheckPasswordHash(inputPass, hash) != nil {
		t.Fatalf("Password and hashed password not the same")
		return
	}
}

func TestJWT(t *testing.T) {

	userID := uuid.New()
	tokenSecret := "TestSecret"
	tokenString, err := MakeJWT(userID, tokenSecret, time.Second)
	if err != nil {
		t.Fatalf("Error creating JWT: %v", err)
		return
	}

	tokenUserID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("Error validating the jwt: %v", err)
		return
	}

	if tokenUserID != userID {
		t.Fatalf("Wrong matching of userIDs")
		return
	}
}
