package main

import (
	"fmt"
	"io/ioutil"
	"time"
)

func main() {
	date := time.Now()
	fDate := date.Format(time.RFC3339)

	filename := fmt.Sprintf("%s.txt", fDate)

	err := ioutil.WriteFile(fmt.Sprintf("./%s", filename), []byte(fmt.Sprintf("date: %s\n\n", fDate)), 0644)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(fmt.Sprintf("Created: %s", filename))
}
