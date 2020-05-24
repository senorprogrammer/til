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
	"sync"
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
editor: ""
targetDirectory: "~/Documents/til"
`
	defaultEditor = "open"
	fileExtension = "md"

	// A custom datetime format that plays nicely with GitHub Pages filename restrictions
	ghFriendlyDateFormat = "2006-01-02T15-04-05"

	/* -------------------- Messages -------------------- */

	errConfigDirCreate    = "could not create the configuration directory"
	errConfigExpandPath   = "could not expand the config directory"
	errConfigFileAssert   = "could not assert the configuration file exists"
	errConfigFileCreate   = "could not create the configuration file"
	errConfigFileWrite    = "could not write the configuration file"
	errConfigPathEmpty    = "the config path cannot be empty"
	errConfigValueRead    = "could not read a required configuration value"
	errNoTitle            = "title must not be blank"
	errTargetDirCreate    = "could not create the target directories"
	errTargetDirFlag      = "multiple target directories defined, no -t value provided"
	errTargetDirUndefined = "target directory is undefined or misconfigured in config"

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
)

// globalConfig holds and makes available all the user-configurable
// settings that are stored in the config file.
// (I know! Friends don't let friends use globals, but since I have
// no friends working on this, there's no one around to stop me)
var globalConfig *config.Config

// ll is a go routine-safe implementation of Logger
// (More globals! This is getting crazy)
var ll *log.Logger

var buildFlag bool
var listFlag bool
var saveFlag bool
var targetDirFlag string

func init() {
	ll = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	flag.BoolVar(&buildFlag, "b", false, "builds the index and tag pages (short-hand)")
	flag.BoolVar(&buildFlag, "build", false, "builds the index and tag pages")

	flag.BoolVar(&listFlag, "l", false, "lists the configured target directories (short-hand)")
	flag.BoolVar(&listFlag, "list", false, "lists the configured target directories")

	flag.BoolVar(&saveFlag, "s", false, "builds, saves, and pushes (short-hand)")
	flag.BoolVar(&saveFlag, "save", false, "builds, saves, and pushes")

	flag.StringVar(&targetDirFlag, "t", "", "specifies the target directory key (short-hand)")
	flag.StringVar(&targetDirFlag, "target", "", "specifies the target directory key")
}

func main() {
	flag.Parse()

	loadConfig()

	/* Flaghandling */
	/* I personally think "flag handling" should be spelled flag-handling
	   but precedence has been set and we will defer to it.
	   According to wiktionary.org, "stick-handling" is correctly spelled
	   "stickhandling", so here we are, abomination enshrined */

	if listFlag {
		listTargetDirectories(globalConfig)
		Victory(statusDone)
	}

	if buildFlag {
		buildContent()
		Victory(statusDone)
	}

	if saveFlag {
		commitMsg := strings.Join(os.Args[2:], " ")

		buildContent()
		save(commitMsg)
		push()
		Victory(statusDone)
	}

	buildTargetDirectory()

	/* Page creation */

	title := parseTitle(targetDirFlag, os.Args)
	if title == "" {
		// Every non-dash argument is considered a part of the title. If there are no arguments, we have no title
		// Can't have a page without a title
		Defeat(errors.New(errNoTitle))
	}

	pagePath := createNewPage(title)

	// Write the pagePath to the console. This makes it easy to know which file we just created
	Info(pagePath)

	// And rebuild the index and tag pages
	buildContent()

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

// getConfigFilePath returns the string path to the configuration file
func getConfigFilePath() (string, error) {
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

	if _, err := os.Stat(cDir); os.IsNotExist(err) {
		err := os.MkdirAll(cDir, os.ModePerm)
		if err != nil {
			Defeat(errors.New(errConfigDirCreate))
		}

		Progress(fmt.Sprintf("created %s", cDir))
	}
}

func makeConfigFile() {
	cPath, err := getConfigFilePath()
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
func readConfigFile() {
	cPath, err := getConfigFilePath()
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

	globalConfig = cfg
}

/* -------------------- Target Directory -------------------- */

// buildTargetDirectory verifies that the target directory, as specified in
// the config file, exists and contains a /docs folder for writing pages to.
// If these directories don't exist, it tries to create them
func buildTargetDirectory() {
	tDir, err := getTargetDir(globalConfig, true)
	if err != nil {
		Defeat(err)
	}

	if _, err := os.Stat(tDir); os.IsNotExist(err) {
		err := os.MkdirAll(tDir, os.ModePerm)
		if err != nil {
			Defeat(errors.New(errTargetDirCreate))
		}
	}
}

// getTargetDir returns the absolute string path to the directory that the
// content will be written to
func getTargetDir(cfg *config.Config, withDocsDir bool) (string, error) {
	docsBit := ""
	if withDocsDir {
		docsBit = "/docs"
	}

	// Target directories are defined in the config file as a map of
	// identifier : target directory
	// Example:
	//		targetDirectories:
	//			a: ~/Documents/blog
	//			b: ~/Documents/notes
	uDirs, err := cfg.Map("targetDirectories")
	if err != nil {
		return "", err
	}

	// config returns a map of [string]interface{} which is helpful on the
	// left side, not so much on the right side. Convert the right to strings
	tDirs := make(map[string]string, len(uDirs))
	for k, dir := range uDirs {
		tDirs[k] = dir.(string)
	}

	// Extracts the dir we want operate against by using the value of the
	// -target flag passed in. If no value was passed in, AND we only have one
	// entry in the map, use that entry. If no value was passed in and there
	// are multiple entries in the map, raise an error because ¯\_(ツ)_/¯
	tDir := ""

	if len(tDirs) == 1 {
		for _, dir := range tDirs {
			tDir = dir
		}
	} else {
		if targetDirFlag == "" {
			return "", errors.New(errTargetDirFlag)
		}

		tDir = tDirs[targetDirFlag]
	}

	if tDir == "" {
		return "", errors.New(errTargetDirUndefined)
	}

	// If we're not using a path relative to the user's home directory,
	// take the config value as a fully-qualified path and just append the
	// name of the write dir to it
	if tDir[0] != '~' {
		return tDir + docsBit, nil
	}

	// We are pathing relative to the home directory, so figure out the
	// absolute path for that
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New(errConfigExpandPath)
	}

	return filepath.Join(dir, tDir[1:], docsBit), nil
}

/* -------------------- Helper functions -------------------- */

func buildContent() {
	pages := loadPages()
	tagMap := buildTagPages(pages)
	buildIndexPage(pages, tagMap)
}

// buildIndexPage creates the main index.md page that is the root of the site
func buildIndexPage(pages []*Page, tagMap *TagMap) {
	Info(statusIdxBuild)

	content := ""

	// Write the tag list into the top of the index
	for _, tag := range tagMap.SortedTagNames() {
		content += fmt.Sprintf(
			"[%s](%s), ",
			tag,
			fmt.Sprintf("./%s", tag),
		)
	}

	// Write the page list into the middle of the page
	content += pagesToHTMLUnorderedList(pages)
	content += "\n"

	// Write the footer content into the bottom of the index
	content += "\n"
	content += footer()

	// And write the file to disk
	tDir, err := getTargetDir(globalConfig, true)
	if err != nil {
		Defeat(err)
	}

	filePath := fmt.Sprintf(
		"%s/index.%s",
		tDir,
		fileExtension,
	)

	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		Defeat(err)
	}

	Progress(filePath)
}

// buildTagPages creates the tag pages, with links to posts tagged with those names
func buildTagPages(pages []*Page) *TagMap {
	Info(statusTagBuild)

	tagMap := NewTagMap(pages)

	var wGroup sync.WaitGroup

	for _, tagName := range tagMap.SortedTagNames() {
		wGroup.Add(1)

		go func(tagName string, ll *log.Logger) {
			defer wGroup.Done()

			content := fmt.Sprintf("## %s\n\n", tagName)

			// Write the page list into the middle of the page
			content += pagesToHTMLUnorderedList(tagMap.PagesFor(tagName))

			// Write the footer content into the bottom of the page
			content += "\n"
			content += footer()

			// And write the file to disk
			tDir, err := getTargetDir(globalConfig, true)
			if err != nil {
				Defeat(err)
			}

			filePath := fmt.Sprintf(
				"%s/%s.%s",
				tDir,
				tagName,
				fileExtension,
			)

			err = ioutil.WriteFile(filePath, []byte(content), 0644)
			if err != nil {
				Defeat(err)
			}

			Progress(filePath)
		}(tagName, ll)
	}

	wGroup.Wait()

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
	tDir, err := getTargetDir(globalConfig, true)
	if err != nil {
		Defeat(err)
	}

	filePath := fmt.Sprintf(
		"%s/%s-%s.%s",
		tDir,
		pathDate,
		strings.ReplaceAll(strings.ToLower(title), " ", "-"),
		fileExtension,
	)

	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		Defeat(err)
	}

	// Tell the OS to open the newly-created page in the editor (as specified in the config)
	// If there's no editor explicitly defined by the user, tell the OS to try and open it
	editor := globalConfig.UString("editor", defaultEditor)
	if editor == "" {
		editor = defaultEditor
	}

	cmd := exec.Command(editor, filePath)
	err = cmd.Run()
	if err != nil {
		Defeat(err)
	}

	return filePath
}

// listTargetDirectories writes the list of target directories in the configuration
// out to the terminal
func listTargetDirectories(cfg *config.Config) {
	dirMap, err := cfg.Map("targetDirectories")
	if err != nil {
		Defeat(err)
	}

	for key, dir := range dirMap {
		Info(fmt.Sprintf("%6s\t%s\n", key, dir.(string)))
	}
}

// loadPages reads the page files from disk (in reverse chronological order) and
// creates Page instances from them
func loadPages() []*Page {
	pages := []*Page{}

	tDir, err := getTargetDir(globalConfig, true)
	if err != nil {
		Defeat(err)
	}

	filePaths, _ := filepath.Glob(
		fmt.Sprintf(
			"%s/*.%s",
			tDir,
			fileExtension,
		),
	)

	for i := len(filePaths) - 1; i >= 0; i-- {
		page := readPage(filePaths[i])
		pages = append(pages, page)
	}

	return pages
}

// pagesToHTMLUnorderedList creates the unordered list of page links that appear
// on the index and tag pages
func pagesToHTMLUnorderedList(pages []*Page) string {
	content := ""
	prevPage := &Page{}

	for _, page := range pages {
		if !page.IsContentPage() {
			continue
		}

		// This breaks the page list up by month
		if prevPage.CreatedMonth() != page.CreatedMonth() {
			content += "\n"
		}

		content += fmt.Sprintf("* %s\n", page.Link())

		prevPage = page
	}

	return content
}

func parseTitle(targetFlag string, args []string) string {
	titleOffset := 3
	if targetFlag == "" {
		titleOffset = 1
	}

	return strings.Title(strings.Join(args[titleOffset:], " "))
}

// push pushes up to the remote git repo
func push() {
	Info(statusRepoPush)

	tDir, err := getTargetDir(globalConfig, false)
	if err != nil {
		Defeat(err)
	}

	r, err := git.PlainOpen(tDir)
	if err != nil {
		Defeat(err)
	}

	err = r.Push(&git.PushOptions{})
	if err != nil {
		Defeat(err)
	}
}

// readPage reads the contents of the page and unmarshals it into the Page struct,
// making the page's internal frontmatter programmatically accessible
func readPage(filePath string) *Page {
	page := new(Page)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		Defeat(err)
	}

	err = frontmatter.Unmarshal(data, page)
	if err != nil {
		Defeat(err)
	}

	page.FilePath = filePath

	return page
}

// https://github.com/go-git/go-git/blob/master/_examples/commit/main.go
func save(commitMsg string) {
	Info(statusRepoSave)

	tDir, err := getTargetDir(globalConfig, false)
	if err != nil {
		Defeat(err)
	}

	r, err := git.PlainOpen(tDir)
	if err != nil {
		Defeat(err)
	}

	w, err := r.Worktree()
	if err != nil {
		Defeat(err)
	}

	_, err = w.Add(".")
	if err != nil {
		Defeat(err)
	}

	defaultCommitMsg, err1 := globalConfig.String("commitMessage")
	defaultCommitEmail, err2 := globalConfig.String("committerEmail")
	defaultCommitName, err3 := globalConfig.String("committerName")
	if err1 != nil || err2 != nil || err3 != nil {
		Defeat(errors.New(errConfigValueRead))
	}

	if commitMsg == "" {
		// The incoming commitMsg is optional (if it is set, it probably came in
		// via command line args on -save). If it isn't set, we use the default
		// from the config file instead
		commitMsg = defaultCommitMsg
	}

	commit, err := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  defaultCommitName,
			Email: defaultCommitEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		Defeat(err)
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		Defeat(err)
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

// Defeat writes out an error message
func Defeat(err error) {
	ll.Fatal(fmt.Sprintf("%s %s", Red("✘"), err.Error()))
}

// Info writes out an informative message
func Info(msg string) {
	ll.Print(fmt.Sprintf("%s %s", Green("->"), msg))
}

// Progress writes out a progress status message
func Progress(msg string) {
	ll.Print(fmt.Sprintf("\t%s %s\n", Blue("->"), msg))
}

// Victory writes out a victorious final message and then expires dramatically
func Victory(msg string) {
	ll.Print(fmt.Sprintf("%s %s", Green("✓"), msg))
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
		filepath.Base(page.FilePath),
	)
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

// PagesFor returns a flattened slice of pages for a given tag name, sorted
// in reverse-chronological order
func (tm *TagMap) PagesFor(tagName string) []*Page {
	pages := []*Page{}
	tags := tm.Get(tagName)

	for _, tag := range tags {
		pages = append(pages, tag.Pages...)

		sort.Slice(pages, func(i, j int) bool {
			return pages[i].CreatedAt().After(pages[j].CreatedAt())
		})
	}

	return pages
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
