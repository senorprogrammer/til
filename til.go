package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
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
		buildIndex()
		os.Exit(0)
	}

	date := time.Now()
	pathDate := date.Format("2006-01-02T15-04-05")

	title := strings.Title(strings.Join(os.Args[1:], " "))
	filepath := fmt.Sprintf("./docs/%s-%s.%s", pathDate, strings.ReplaceAll(strings.ToLower(title), " ", "-"), fileExtension)

	// Front matter lives at the top of the generated file and contains bits of info about the file
	// This is loosely based on the format Hugo uses
	frontMatter := fmt.Sprintf(
		"---\ndate: %s\ntitle: %s\ntags: %s\n---\n\n",
		date.Format(time.RFC3339),
		title,
		"",
	)

	content := frontMatter + fmt.Sprintf("# %s\n\n\n", title)

	// Write out the stub file, explode if we can't do that
	err := ioutil.WriteFile(fmt.Sprintf("%s", filepath), []byte(content), 0644)
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

func buildIndex() {
	content := "A collection of things\n\n"

	files, _ := filepath.Glob("./docs/*.md")

	// Loop over the files in reverse, which puts them in reverse-chronological order
	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]

		ma := readFrontMatter(file)

		if ma.IsValid() {
			postDate, _ := time.Parse(time.RFC3339, ma.Date)
			prettyDate := postDate.Format("Jan 02, 2006")

			content += fmt.Sprintf("* <code>%s</code> [%s](%s)\n", prettyDate, ma.Title, strings.Replace(file, "docs/", "", -1))
		}
	}

	content += fmt.Sprintf("\n")

	content += fmt.Sprintf("<sup><sub>generated %s</sub></sup>\n", time.Now().Format("2 Jan 2006 15:04:05"))

	err := ioutil.WriteFile("./docs/index.md", []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func readFrontMatter(filePath string) *Matter {
	ma := new(Matter)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	err = frontmatter.Unmarshal(([]byte)(data), ma)
	if err != nil {
		log.Fatal(err)
	}

	return ma
}

/* -------------------- Types -------------------- */

// Matter represents the frontmatter stored in the top of each markdown page
type Matter struct {
	Date    string `yaml:"date"`
	Title   string `yaml:"title"`
	Tags    string `yaml:"tags"`
	Content string `fm:"content" yaml:"-"`
}

// IsValid returns true if the page is a valid entry page, false if it is not
func (ma *Matter) IsValid() bool {
	return ma.Title != ""
}
