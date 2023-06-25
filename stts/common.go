package main

import (
	"errors"
	"fmt"
	"os"
)

func fileExists(file string) bool {
	_, err := os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func elInSlice(sl []string, el string) bool {
	for _, s := range sl {
		if s == el {
			return true
		}
	}
	return false
}

func errExit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
