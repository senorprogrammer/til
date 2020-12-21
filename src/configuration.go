package src

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/olebedev/config"
)

const (
	defaultConfig = `--- 
commitMessage: "build, save, push"
committerEmail: test@example.com
committerName: "TIL Autobot"
editor: ""
targetDirectory: 
	a: "~/Documents/tilblog"
`

	tilConfigDir  = "~/.config/til/"
	tilConfigFile = "config.yml"
)

const (
	errConfigDirCreate  = "could not create the configuration directory"
	errConfigExpandPath = "could not expand the config directory"
	errConfigFileAssert = "could not assert the configuration file exists"
	errConfigFileCreate = "could not create the configuration file"
	errConfigFileWrite  = "could not write the configuration file"
	errConfigPathEmpty  = "config path cannot be empty"
)

// GlobalConfig holds and makes available all the user-configurable
// settings that are stored in the config file.
// (I know! Friends don't let friends use globals, but since I have
// no friends working on this, there's no one around to stop me)
var GlobalConfig *config.Config

// Config handles all things to do with configuration
type Config struct{}

// Load reads the configuration file
func (c *Config) Load() {
	makeConfigDir()
	makeConfigFile()

	GlobalConfig = readConfigFile()
}

// getConfigDir returns the string path to the directory that should
// contain the configuration file.
// It tries to be XDG-compatible
func getConfigDir() (string, error) {
	cDir := os.Getenv("XDG_CONFIG_HOME")
	if cDir == "" {
		cDir = tilConfigDir
	}

	// If the user hasn't changed the default path then we expect it to start
	// with a tilde (the user's home), and we need to turn that into an
	// absolute path. If it does not start with a '~' then we assume the
	// user has set their $XDG_CONFIG_HOME to something specific, and we
	// do not mess with it (because doing so makes the archlinux people
	// very cranky)
	if cDir[0] != '~' {
		return cDir, nil
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New(errConfigExpandPath)
	}

	cDir = filepath.Join(dir, cDir[1:])

	if cDir == "" {
		return "", errors.New(errConfigPathEmpty)
	}

	return cDir, nil
}

// GetConfigFilePath returns the string path to the configuration file
func GetConfigFilePath() (string, error) {
	cDir, err := getConfigDir()
	if err != nil {
		return "", err
	}

	if cDir == "" {
		return "", errors.New(errConfigPathEmpty)
	}

	return fmt.Sprintf("%s/%s", cDir, tilConfigFile), nil
}

func makeConfigDir() {
	cDir, err := getConfigDir()
	if err != nil {
		Defeat(err)
	}

	if cDir == "" {
		Defeat(errors.New(errConfigPathEmpty))
	}

	if _, err := os.Stat(cDir); os.IsNotExist(err) {
		err := os.MkdirAll(cDir, os.ModePerm)
		if err != nil {
			Defeat(errors.New(errConfigDirCreate))
		}

		Progress(fmt.Sprintf("created %s", cDir))
	}
}

func makeConfigFile() {
	cPath, err := GetConfigFilePath()
	if err != nil {
		Defeat(err)
	}

	if cPath == "" {
		Defeat(errors.New(errConfigPathEmpty))
	}

	_, err = os.Stat(cPath)

	if err != nil {
		// Something went wrong trying to find the config file.
		// Let's see if we can figure out what happened
		if os.IsNotExist(err) {
			// Ah, the config file does not exist, which is probably fine
			_, err = os.Create(cPath)
			if err != nil {
				// That was not fine
				Defeat(errors.New(errConfigFileCreate))
			}

		} else {
			// But wait, it's some kind of other error. What kind?
			// I dunno, but it's probably bad so die
			Defeat(err)
		}
	}

	// Let's double-check that the file's there now
	fileInfo, err := os.Stat(cPath)
	if err != nil {
		Defeat(errors.New(errConfigFileAssert))
	}

	// Write the default config, but only if the file is empty.
	// Don't want to stop on any non-default values the user has written in there
	if fileInfo.Size() == 0 {
		if ioutil.WriteFile(cPath, []byte(defaultConfig), 0600) != nil {
			Defeat(errors.New(errConfigFileWrite))
		}

		Progress(fmt.Sprintf("created %s", cPath))
	}
}

// readConfigFile reads the contents of the config file and jams them
// into the global config variable
func readConfigFile() *config.Config {
	cPath, err := GetConfigFilePath()
	if err != nil {
		Defeat(err)
	}

	if cPath == "" {
		Defeat(errors.New(errConfigPathEmpty))
	}

	cfg, err := config.ParseYamlFile(cPath)
	if err != nil {
		Defeat(err)
	}

	return cfg
}
