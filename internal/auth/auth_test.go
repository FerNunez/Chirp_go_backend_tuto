package auth

import (
	"fmt"
	"net/http"
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

func TestGetBearerToken(t *testing.T) {

	expectedToken := "WhatEver_dude"

	header := http.Header{}
	expectedString := fmt.Sprintf("Bearer %v", expectedToken)
	header.Set("Authorization", expectedString)

	token, err := GetBearerToken(header)
	if err != nil {
		t.Fatalf("Error while getting bearer token: %v", err)
		return
	}

	if token != expectedToken {
		t.Fatalf("Expected: %v and gotten token: %v are different", expectedToken, token)
	}
}

func TestGetAPIKey(t *testing.T) {

	expectedToken := "WhatEver_dude"

	header := http.Header{}
	expectedString := fmt.Sprintf("ApiKey %v", expectedToken)
	header.Set("Authorization", expectedString)

	token, err := GetAPIKey(header)
	if err != nil {
		t.Fatalf("Error while getting api key token: %v", err)
		return
	}

	if token != expectedToken {
		t.Fatalf("Expected: %v and gotten token: %v are different", expectedToken, token)
	}
}
