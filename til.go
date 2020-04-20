package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	editor        = "mvim"
	fileExtension = "md"
)

func main() {
	if len(os.Args[1:]) < 1 {
		fmt.Println("Must have a title")
		os.Exit(1)
	}

	date := time.Now().Format(time.RFC3339)
	title := strings.Join(os.Args[1:], " ")
	filepath := fmt.Sprintf("%s-%s.%s", date, strings.ReplaceAll(strings.ToLower(title), " ", "-"), fileExtension)

	metadata := fmt.Sprintf(
		"---\ndate: %s\ntitle: %s\n---\n\n\n",
		date,
		strings.Title(title),
	)

	err := ioutil.WriteFile(fmt.Sprintf("./%s", filepath), []byte(metadata), 0644)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(editor, filepath)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(filepath)
	os.Exit(0)
}
