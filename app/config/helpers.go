package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rantanevich/homepage/app/web"
)

func findIcon(userIcons, staticIcons []string, path string) string {
	if path == "" {
		return "/static/icons/no-icon.svg"
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	ext := filepath.Ext(path)
	if ext != "" {
		if slices.Contains(userIcons, path) {
			return filepath.Join("/icons/", path)
		}

		if slices.Contains(staticIcons, path) {
			return filepath.Join("/static/icons/", path)
		}
		return "/static/icons/no-icon.svg"
	}

	for _, ext := range []string{".svg", ".png"} {
		fname := path + ext
		if slices.Contains(userIcons, fname) {
			return filepath.Join("/icons/", fname)
		}

		if slices.Contains(staticIcons, fname) {
			return filepath.Join("/static/icons/", fname)
		}
	}
	return "/static/icons/no-icon.svg"
}

func getStaticIcons() ([]string, error) {
	var icons []string

	files, err := web.WebFS.ReadDir("static/icons")
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
