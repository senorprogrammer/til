package src

import (
	"fmt"
	"log"
	"os"
)

// LL is a go routine-safe implementation of Logger
// (More globals! This is getting crazy)
var LL *log.Logger

// Defeat writes out an error message
func Defeat(err error) {
	LL.Fatal(fmt.Sprintf("%s %s", Red("✘"), err.Error()))
}

// Info writes out an informative message
func Info(msg string) {
	LL.Print(fmt.Sprintf("%s %s", Green("->"), msg))
}

// Progress writes out a progress status message
func Progress(msg string) {
	LL.Print(fmt.Sprintf("\t%s %s\n", Blue("->"), msg))
}

// Victory writes out a victorious final message and then expires dramatically
func Victory(msg string) {
	LL.Print(fmt.Sprintf("%s %s", Green("✓"), msg))
	os.Exit(0)
}
