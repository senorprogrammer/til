package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/olebedev/config"
)

const (
	// Configuration
	tilConfigDir  = "~/.config/til/"
	tilConfigFile = "config.yml"

	defaultConfig = `--- 
commitMessage: "build, save, push"
committerEmail: test@example.com
committerName: "TIL Autobot"
editor: "mvim"
`

	fileExtension = "md"

	// A custom datetime format that plays nicely with GitHub Pages filename restrictions
	ghFriendlyDateFormat = "2006-01-02T15-04-05"

	/* -------------------- Messages -------------------- */

	errConfigDirCreate  = "could not create the configuration directory"
	errConfigExpandPath = "could not expand the config directory"
	errConfigFileAssert = "could not assert the configuration file exists"
	errConfigFileCreate = "could not create the configuration file"
	errConfigFileWrite  = "could not write the configuration file"
	errConfigValueRead  = "could not read a required configuration value"
	errNoTitle          = "title must not be blank"

	statusDone     = "done"
	statusIdxBuild = "building index page"
	statusRepoPush = "pushing to remote"
	statusRepoSave = "saving uncommitted files"
	statusTagBuild = "building tag pages"
)

var (
	// Blue writes blue text
	Blue = Colour("\033[1;36m%s\033[0m")

	// Green writes green text
	Green = Colour("\033[1;32m%s\033[0m")

	// Red writes red text
	Red = Colour("\033[1;31m%s\033[0m")

	// Yellow writes yellow text
	Yellow = Colour("\033[1;33m%s\033[0m")
)

// globalConfig holds and makes available all the user-configurable
// settings that are stored in the config file.
// (I know! Friends don't let friends use globals, but since I have
// no friends working on this, there's no one around to stop me)
var globalConfig *config.Config

var buildFlag bool
var saveFlag bool

func init() {
	log.SetOutput(os.Stderr)
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.BoolVar(&buildFlag, "b", false, "builds the index and tag pages (short-hand)")
	flag.BoolVar(&buildFlag, "build", false, "builds the index and tag pages")

	flag.BoolVar(&saveFlag, "s", false, "builds, saves, and pushes (short-hand)")
	flag.BoolVar(&saveFlag, "save", false, "builds, saves, and pushes")
}

func main() {
	loadConfig()

	/* Flags */

	flag.Parse()

	if buildFlag {
		build()
		Victory(statusDone)
	}

	if saveFlag {
		commitMsg := strings.Join(os.Args[2:], " ")

		build()
		save(commitMsg)
		push()
		Victory(statusDone)
	}

	/* Page creation */

	// Every non-dash argument is considered a part of the title. If there are no arguments, we have no title
	// Can't have a page without a title
	if len(os.Args[1:]) < 1 {
		Fail(errors.New(errNoTitle))
	}

	title := strings.Title(strings.Join(os.Args[1:], " "))

	pagePath := createNewPage(title)

	// Write the pagePath to the console. This makes it easy to know which file we just created
	log.Print(fmt.Sprintf("%s %s", Green("->"), pagePath))

	// And rebuild the index and tag pages
	build()

	Victory(statusDone)
}

/* -------------------- Configuration -------------------- */

func loadConfig() {
	makeConfigDir()
	makeConfigFile()
	readConfigFile()
}

// getConfigDir returns the string path to the directory that should
// contain the configuration file.
// It tries to be XDG-compatible
func getConfigDir() string {
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
		return cDir
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		Fail(errors.New(errConfigExpandPath))
	}

	return filepath.Join(dir, cDir[1:])
}

// getConfigPath returns the string path to the configuration file
func getConfigPath() string {
	cDir := getConfigDir()
	cPath := fmt.Sprintf("%s/%s", cDir, tilConfigFile)

	return cPath
}

func makeConfigDir() {
	cDir := getConfigDir()

	if _, err := os.Stat(cDir); os.IsNotExist(err) {
		err := os.MkdirAll(cDir, os.ModePerm)
		if err != nil {
			Fail(errors.New(errConfigDirCreate))
		}
	}
}

func makeConfigFile() {
	cPath := getConfigPath()

	fileInfo, err := os.Stat(cPath)

	if err != nil {
		// Something went wrong trying to find the config file.
		// Let's see if we can figure out what happened
		if os.IsNotExist(err) {
			// Ah, the config file does not exist, which is probably fine
			_, err = os.Create(cPath)
			if err != nil {
				// That was not fine
				Fail(errors.New(errConfigFileCreate))
			}
		} else {
			// But wait, it's some kind of other error. What kind?
			// I dunno, but it's probably bad so die
			Fail(err)
		}
	}

	// Let's double-check that the file's there now
	fileInfo, err = os.Stat(cPath)
	if err != nil {
		Fail(errors.New(errConfigFileAssert))
	}

	// Write the default config, but only if the file is empty.
	// Don't want to stop on any non-default values the user has written in there
	if fileInfo.Size() == 0 {
		if ioutil.WriteFile(cPath, []byte(defaultConfig), 0600) != nil {
			Fail(errors.New(errConfigFileWrite))
		}
	}
}

// readConfigFile reads the contents of the config file and jams them
// into the global config variable
func readConfigFile() {
	cPath := getConfigPath()

	cfg, err := config.ParseYamlFile(cPath)
	if err != nil {
		Fail(err)
	}

	globalConfig = cfg
}

/* -------------------- Helper functions -------------------- */

func build() {
	pages := loadPages()
	tagMap := buildTagPages(pages)
	buildIndexPage(pages, tagMap)
}

func buildIndexPage(pages []*Page, tagMap *TagMap) {
	Info(statusIdxBuild)

	content := ""
	prevPage := &Page{}

	// Write the tag list into the top of the index
	for _, tag := range tagMap.SortedTagNames() {
		content += fmt.Sprintf(
			"[%s](%s), ",
			tag,
			fmt.Sprintf("./%s", tag),
		)
	}

	// Write the page list into the middle of the index
	// This is sorted by month, in reverse-chronological order
	for _, page := range pages {
		if !page.IsContentPage() {
			continue
		}

		// This breaks the page list up by month
		if prevPage.CreatedMonth() != page.CreatedMonth() {
			content += "\n\n"
		}

		content += fmt.Sprintf("* %s\n", page.Link())

		prevPage = page
	}

	content += fmt.Sprintf("\n")

	// Write the footer content into the bottom of the index
	content += fmt.Sprintf("\n")
	content += fmt.Sprintf("\n")
	content += footer()

	// And write the file to disk
	err := ioutil.WriteFile(fmt.Sprintf("./docs/index.%s", fileExtension), []byte(content), 0644)
	if err != nil {
		Fail(err)
	}
}

// buildTagPages creates the tag pages, with links to posts tagged with those names
func buildTagPages(pages []*Page) *TagMap {
	Info(statusTagBuild)

	tagMap := NewTagMap(pages)

	for _, tagName := range tagMap.SortedTagNames() {
		go func(tagName string) {
			content := fmt.Sprintf("## %s\n\n", tagName)

			for _, tag := range tagMap.Get(tagName) {
				for _, page := range tag.Pages {
					if page.IsContentPage() {
						content += fmt.Sprintf("* %s\n", page.Link())
					}
				}
			}

			// Write the footer content into the bottom of the page
			content += fmt.Sprintf("\n")
			content += footer()

			// And write the file to disk
			fileName := fmt.Sprintf("./docs/%s.%s", tagName, fileExtension)

			err := ioutil.WriteFile(fileName, []byte(content), 0644)
			if err != nil {
				Fail(err)
			}

			log.Print(fmt.Sprintf("%s %s\n", Blue("\t->"), fileName))
		}(tagName)
	}

	return tagMap
}

func createNewPage(title string) string {
	date := time.Now()
	pathDate := date.Format(ghFriendlyDateFormat)

	// Front matter lives at the top of the generated file and contains bits of info about the page
	// This is loosely based on the format Hugo uses
	frontMatter := fmt.Sprintf(
		"---\ndate: %s\ntitle: %s\ntags: %s\n---\n\n",
		date.Format(time.RFC3339),
		title,
		"",
	)

	content := frontMatter + fmt.Sprintf("# %s\n\n\n", title)

	// Write out the stub file, explode if we can't do that
	filePath := fmt.Sprintf("./docs/%s-%s.%s", pathDate, strings.ReplaceAll(strings.ToLower(title), " ", "-"), fileExtension)

	err := ioutil.WriteFile(fmt.Sprintf("%s", filePath), []byte(content), 0644)
	if err != nil {
		Fail(err)
	}

	// Tell the OS to open the newly-created page in the editor (as specified in the config)
	editor, err1 := globalConfig.String("editor")
	if err1 != nil {
		Fail(err1)
	}

	cmd := exec.Command(editor, filePath)
	err = cmd.Run()
	if err != nil {
		Fail(err)
	}

	return filePath
}

// loadPages reads the page files from disk (in reverse chronological order) and
// creates Page instances from them
func loadPages() []*Page {
	pages := []*Page{}

	filePaths, _ := filepath.Glob(fmt.Sprintf("./docs/*.%s", fileExtension))

	for i := len(filePaths) - 1; i >= 0; i-- {
		page := readPage(filePaths[i])
		pages = append(pages, page)
	}

	return pages
}

// push pushes up to the remote git repo
func push() {
	Info(statusRepoPush)

	r, err := git.PlainOpen(".")
	if err != nil {
		Fail(err)
	}

	err = r.Push(&git.PushOptions{})
	if err != nil {
		Fail(err)
	}
}

// readPage reads the contents of the page and unmarshals it into the Page struct,
// making the page's internal frontmatter programmatically accessible
func readPage(filePath string) *Page {
	page := new(Page)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		Fail(err)
	}

	err = frontmatter.Unmarshal(([]byte)(data), page)
	if err != nil {
		Fail(err)
	}

	page.FilePath = filePath

	return page
}

// https://github.com/go-git/go-git/blob/master/_examples/commit/main.go
func save(commitMsg string) {
	Info(statusRepoSave)

	r, err := git.PlainOpen(".")
	if err != nil {
		Fail(err)
	}

	w, err := r.Worktree()
	if err != nil {
		Fail(err)
	}

	_, err = w.Add(".")
	if err != nil {
		Fail(err)
	}

	defaultCommitMsg, err1 := globalConfig.String("commitMessage")
	defaultCommitEmail, err2 := globalConfig.String("committerEmail")
	defaultCommitName, err3 := globalConfig.String("committerName")
	if err1 != nil || err2 != nil || err3 != nil {
		Fail(errors.New(errConfigValueRead))
	}

	if commitMsg == "" {
		// The incoming commitMsg is optional (if it is set, it probably came in
		// via command line args on -save). If it isn't set, we use the default
		// from the config file instead
		commitMsg = defaultCommitMsg
	}

	commit, err := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  defaultCommitEmail,
			Email: defaultCommitName,
			When:  time.Now(),
		},
	})
	if err != nil {
		Fail(err)
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		Fail(err)
	}

	Info(fmt.Sprintf("committed with '%s' (%.7s)", obj.Message, obj.Hash.String()))
}

/* -------------------- More Helper Functions -------------------- */

// Colour returns a function that defines a printable colour string
func Colour(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}

// Fail writes out an error message
func Fail(err error) {
	log.Fatal(fmt.Sprintf("%s %s", Red("x"), err.Error()))
}

// Info writes out an informative message
func Info(msg string) {
	log.Print(fmt.Sprintf("%s %s", Green("->"), msg))
}

// Victory writes out a victorious final message and then expires dramatically
func Victory(msg string) {
	log.Print(msg)
	os.Exit(0)
}

func footer() string {
	return fmt.Sprintf(
		"<sup><sub>generated %s by <a href='https://github.com/senorprogrammer/til'>til</a></sub></sup>\n",
		time.Now().Format("2 Jan 2006 15:04:05"),
	)
}

/* -------------------- Page -------------------- */

// Page represents a TIL page
type Page struct {
	Content  string `fm:"content" yaml:"-"`
	Date     string `yaml:"date"`
	FilePath string `yaml:"filepath"`
	TagsStr  string `yaml:"tags"`
	Title    string `yaml:"title"`
}

// CreatedAt returns a time instance representing when the page was created
func (page *Page) CreatedAt() time.Time {
	date, err := time.Parse(time.RFC3339, page.Date)
	if err != nil {
		return time.Time{}
	}

	return date
}

// CreatedMonth returns the month the page was created
func (page *Page) CreatedMonth() time.Month {
	if page.CreatedAt().IsZero() {
		return 0
	}

	return page.CreatedAt().Month()
}

// IsContentPage returns true if the page is a valid entry page, false if it is not
func (page *Page) IsContentPage() bool {
	return page.Title != ""
}

// Link returns a link string suitable for embedding in a Markdown page
func (page *Page) Link() string {
	return fmt.Sprintf(
		"<code>%s</code> [%s](%s)",
		page.PrettyDate(),
		page.Title,
		strings.Replace(page.FilePath, "docs/", "", -1))
}

// PrettyDate returns a human-friendly representation of the CreatedAt date
func (page *Page) PrettyDate() string {
	return page.CreatedAt().Format("Jan 02, 2006")
}

// Tags returns a slice of tags assigned to this page
func (page *Page) Tags() []*Tag {
	tags := []*Tag{}

	names := strings.Split(page.TagsStr, ",")
	for _, name := range names {
		tags = append(tags, NewTag(name, page))
	}

	return tags
}

/* -------------------- Tag -------------------- */

// Tag represents a page tag (e.g.: linux, zombies)
type Tag struct {
	Name  string
	Pages []*Page
}

// NewTag creates and returns an instance of Tag
func NewTag(name string, page *Page) *Tag {
	tag := &Tag{
		Name:  strings.TrimSpace(name),
		Pages: []*Page{page},
	}

	return tag
}

// AddPage adds a page to the list of pages
func (tag *Tag) AddPage(page *Page) {
	tag.Pages = append(tag.Pages, page)
}

// IsValid returns true if this is a valid tag, false if it is not
func (tag *Tag) IsValid() bool {
	return tag.Name != ""
}

/* -------------------- TagMap -------------------- */

// TagMap is a map of tag name to Tag instance
type TagMap struct {
	Tags map[string][]*Tag
}

// NewTagMap creates and returns an instance of TagMap
func NewTagMap(pages []*Page) *TagMap {
	tm := &TagMap{
		Tags: make(map[string][]*Tag),
	}

	tm.BuildFromPages(pages)

	return tm
}

// Add adds a Tag instance to the map
func (tm *TagMap) Add(tag *Tag) {
	if !tag.IsValid() {
		return
	}

	tm.Tags[tag.Name] = append(tm.Tags[tag.Name], tag)
}

// BuildFromPages populates the tag map from a slice of Page instances
func (tm *TagMap) BuildFromPages(pages []*Page) {
	for _, page := range pages {
		for _, tag := range page.Tags() {
			tm.Add(tag)
		}
	}
}

// Get returns the tags for a given tag name
func (tm *TagMap) Get(name string) []*Tag {
	return tm.Tags[name]
}

// Len returns the number of tags in the map
func (tm *TagMap) Len() int {
	return len(tm.Tags)
}

// SortedTagNames returns the tag names in alphabetical order
func (tm *TagMap) SortedTagNames() []string {
	tagArr := make([]string, tm.Len())
	i := 0

	for tag := range tm.Tags {
		tagArr[i] = tag
		i++
	}

	sort.Strings(tagArr)

	return tagArr
}
