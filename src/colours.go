package src

import "fmt"

var (
	// Blue writes blue text
	Blue = Colour("\033[1;36m%s\033[0m")

	// Green writes green text
	Green = Colour("\033[1;32m%s\033[0m")

	// Red writes red text
	Red = Colour("\033[1;31m%s\033[0m")
)

// Colour returns a function that defines a printable colour string
func Colour(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
