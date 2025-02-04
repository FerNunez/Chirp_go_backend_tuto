package utils

import "strings"

var bannedWords = []string{"kerfuffle", "sharbert", "fornax"}

func cleanProfane(input string) string {
	output := []string{}

	for _, word := range strings.Fields(input) {
		for _, banned := range bannedWords {
			if banned == strings.ToLower(word) {
				word = "****"
			}
		}
		output = append(output, word)
	}
	return strings.Join(output, " ")
}
