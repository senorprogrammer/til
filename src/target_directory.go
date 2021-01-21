package src

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/olebedev/config"
)

const (
	errTargetDirCreate    = "could not create the target directories"
	errTargetDirFlag      = "multiple target directories defined, no -t value provided"
	errTargetDirUndefined = "target directory is undefined or misconfigured in config"
)

// BuildTargetDirectory verifies that the target directory, as specified in
// the config file, exists and contains a /docs folder for writing pages to.
// If these directories don't exist, it tries to create them
func BuildTargetDirectory() {
	tDir, err := GetTargetDir(GlobalConfig, "", true)
	if err != nil {
		log.Printf("BuildTargetDirectory got error: %v", err)
		return
	}

	if _, err := os.Stat(tDir); os.IsNotExist(err) {
		err := os.MkdirAll(tDir, os.ModePerm)
		if err != nil {
			Defeat(errors.New(errTargetDirCreate))
		}
	}
}

// GetTargetDir returns the absolute string path to the directory that the
// content will be written to
func GetTargetDir(cfg *config.Config, targetDirFlag string, withDocsDir bool) (string, error) {
	docsBit := ""
	if withDocsDir {
		docsBit = "/docs"
	}

	// Target directories are defined in the config file as a map of
	// identifier : target directory
	// Example:
	//		targetDirectories:
	//			a: ~/Documents/blog
	//			b: ~/Documents/notes
	uDirs, err := cfg.Map("targetDirectories")
	if err != nil {
		return "", err
	}

	// config returns a map of [string]interface{} which is helpful on the
	// left side, not so much on the right side. Convert the right to strings
	tDirs := make(map[string]string, len(uDirs))
	for k, dir := range uDirs {
		tDirs[k] = dir.(string)
	}

	// Extracts the dir we want operate against by using the value of the
	// -target flag passed in. If no value was passed in, AND we only have one
	// entry in the map, use that entry. If no value was passed in and there
	// are multiple entries in the map, raise an error because ¯\_(ツ)_/¯
	tDir := ""

	if len(tDirs) == 1 {
		for _, dir := range tDirs {
			tDir = dir
		}
	} else {
		if targetDirFlag == "" {
			return "", errors.New(errTargetDirFlag)
		}

		tDir = tDirs[targetDirFlag]
	}

	if tDir == "" {
		return "", errors.New(errTargetDirUndefined)
	}

	// If we're not using a path relative to the user's home directory,
	// take the config value as a fully-qualified path and just append the
	// name of the write dir to it
	if tDir[0] != '~' {
		return tDir + docsBit, nil
	}

	// We are pathing relative to the home directory, so figure out the
	// absolute path for that
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New(errConfigExpandPath)
	}

	return filepath.Join(dir, tDir[1:], docsBit), nil
}
