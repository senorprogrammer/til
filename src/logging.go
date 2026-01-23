package src

import (
	"log"
	"os"
)

// LL is a go routine-safe implementation of Logger
// (More globals! This is getting crazy)
var LL *log.Logger

// Defeat writes out an error message
func Defeat(err error) {
	if LL == nil {
		log.Fatalf("%s %s", Red("✘"), err.Error())
		return
	}
	LL.Fatalf("%s %s", Red("✘"), err.Error())
}

// Info writes out an informative message
func Info(msg string) {
	LL.Printf("%s %s", Green("->"), msg)
}

// Progress writes out a progress status message
func Progress(msg string) {
	LL.Printf("\t%s %s\n", Blue("->"), msg)
}

// Victory writes out a victorious final message and then expires dramatically
func Victory(msg string) {
	LL.Printf("%s %s", Green("✓"), msg)
	os.Exit(0)
}
