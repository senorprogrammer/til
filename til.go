package main

import (
	"fmt"
	"io/ioutil"
	"time"
)

func main() {
	date := time.Now()
	fDate := date.Format(time.RFC3339)

	extension := "md"
	filename := fmt.Sprintf("%s.%s", fDate, extension)

	err := ioutil.WriteFile(fmt.Sprintf("./%s", filename), []byte(fmt.Sprintf("date: %s\n\n", fDate)), 0644)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(fmt.Sprintf("Created: %s", filename))
}
