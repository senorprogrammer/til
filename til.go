package main

import (
	"flag"
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
	// Every argument is considered a part of the title. If there are no arguments, we have no title
	// Cannot have a file without a title
	if len(os.Args[1:]) < 1 {
		fmt.Println("Must have a title")
		os.Exit(1)
	}

	boolPtr := flag.Bool("build", false, "builds the index and tag pages")
	flag.Parse()
	if *boolPtr {
		fmt.Println("building pages....")
		os.Exit(0)
	}

	date := time.Now().Format(time.RFC3339)
	title := strings.Title(strings.Join(os.Args[1:], " "))
	tags := ""
	filepath := fmt.Sprintf("%s-%s.%s", date, strings.ReplaceAll(strings.ToLower(title), " ", "-"), fileExtension)

	// Front matter lives at the top of the generated file and contains bits of info about the file
	// This is loosely based on the format Hugo uses
	frontMatter := fmt.Sprintf(
		"---\ndate: %s\ntitle: %s\ntags: %s\n---\n\n",
		date,
		title,
		tags,
	)

	content := frontMatter + fmt.Sprintf("# %s\n\n\n", title)

	// Write out the stub file, explode if we can't do that
	err := ioutil.WriteFile(fmt.Sprintf("./%s", filepath), []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// And open the file for editing, exploding if we can't do that
	cmd := exec.Command(editor, filepath)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Dump the filepath because this makes it easy to see which file we just created (and delete it, when debugging and testing)
	fmt.Println(filepath)
	os.Exit(0)
}
