package main

import (
	"errors"
	"os"
)

func fileExists(file string) bool {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}
