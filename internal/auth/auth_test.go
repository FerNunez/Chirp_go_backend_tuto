package auth

import (
	"testing"
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
