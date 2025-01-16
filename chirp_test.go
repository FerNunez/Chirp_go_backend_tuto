package main

import (
	"testing"
)

func TestCleanProfaneGood(t *testing.T) {

	input := "This is a test string with no profane"
	expected := "This is a test string with no profane"
	result := cleanProfane(input)

	if result != expected {
		t.Fatalf("cleanProfane(input) `%v` not equal to expected `%v`", result, expected)
	}

}

func TestCleanProfaneEvil(t *testing.T) {

	input := "This is a kerfuffle string with SharBert no profane FORNAX"
	expected := "This is a **** string with **** no profane ****"
	result := cleanProfane(input)

	if result != expected {
		t.Fatalf("cleanProfane(input) `%v` not equal to expected `%v`", result, expected)
	}

}

func TestCleanProfaneSigns(t *testing.T) {

	input := "This is a kerfuffle! string with SharBert! no profane"
	expected := "This is a kerfuffle! string with SharBert! no profane"
	result := cleanProfane(input)

	if result != expected {
		t.Fatalf("cleanProfane(input) `%v` not equal to expected `%v`", result, expected)
	}

}
