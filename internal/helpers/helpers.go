package helpers

import (
	"errors"
	"log"
)

// HandleErr handles error for lambda functions.
// It logs the error and returns the correct lambda handler format
// e is the error string to log and return
func HandleErr(e string) (string, error) {
	log.Println(e)
	return "", errors.New(e)
}
