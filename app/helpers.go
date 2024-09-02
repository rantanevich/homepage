package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
)

func getStaticIcons() ([]string, error) {
	var icons []string

	files, err := webFS.ReadDir("web/static/icons")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		icons = append(icons, file.Name())
	}
	return icons, nil
}

func getUserIcons(path string) ([]string, error) {
	var icons []string

	files, err := os.ReadDir(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return icons, nil
		}
		return nil, err
	}

	for _, file := range files {
		icons = append(icons, file.Name())
	}
	return icons, nil
}

func fatal(err error) {
	if err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
}
