package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/olebedev/config"
	"github.com/senorprogrammer/til/pages"
	"github.com/senorprogrammer/til/src"
)

const (
	defaultCommitMsg = "build, save, push"

	defaultEditor = "open"

	/* -------------------- Messages -------------------- */

	errConfigValueRead = "could not read a required configuration value"
	errNoTitle         = "title must not be blank"

	statusDone     = "done"
	statusIdxBuild = "building index page"
	statusRepoPush = "pushing to remote"
	statusRepoSave = "saving uncommitted files"
	statusTagBuild = "building tag pages"
)

var (
	buildFlag     bool
	listFlag      bool
	saveFlag      bool
	targetDirFlag string
)

func init() {
	src.LL = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	flag.BoolVar(&buildFlag, "b", false, "builds the index and tag pages (short-hand)")
	flag.BoolVar(&buildFlag, "build", false, "builds the index and tag pages")

	flag.BoolVar(&listFlag, "l", false, "lists the configured target directories (short-hand)")
	flag.BoolVar(&listFlag, "list", false, "lists the configured target directories")

	flag.BoolVar(&saveFlag, "s", false, "builds, saves, and pushes (short-hand)")
	flag.BoolVar(&saveFlag, "save", false, "builds, saves, and pushes")

	flag.StringVar(&targetDirFlag, "t", "", "specifies the target directory key (short-hand)")
	flag.StringVar(&targetDirFlag, "target", "", "specifies the target directory key")
}

/* -------------------- Main -------------------- */

func main() {
	flag.Parse()

	cnf := &src.Config{}
	cnf.Load()

	/* Flaghandling */
	/* I personally think "flag handling" should be spelled flag-handling
	   but precedence has been set and we will defer to it.
	   According to wiktionary.org, "stick-handling" is correctly spelled
	   "stickhandling", so here we are, abomination enshrined */

	if listFlag {
		listTargetDirectories(src.GlobalConfig)
		src.Victory(statusDone)
	}

	if buildFlag {
		buildContent()
		src.Victory(statusDone)
	}

	if saveFlag {
		commitMsg := determineCommitMessage(src.GlobalConfig, os.Args)

		buildContent()
		save(commitMsg)
		push()
		src.Victory(statusDone)
	}

	src.BuildTargetDirectory()

	/* Page creation */

	title := parseTitle(targetDirFlag, os.Args)
	if title == "" {
		// Every non-dash argument is considered a part of the title. If there are no arguments, we have no title
		// Can't have a page without a title
		src.Defeat(errors.New(errNoTitle))
	}

	createNewPage(title)

	src.Victory(statusDone)
}

/* -------------------- Helper functions -------------------- */

func buildContent() {
	pages := loadPages()
	buildContentPages(pages)
	tagMap := buildTagPages(pages)

	buildIndexPage(pages, tagMap)
}

// buildContentPages loops through all the pages and tells them to save themselves
// to disk. This process writes any auto-generated content into the pages
func buildContentPages(pages []*pages.Page) {
	for _, page := range pages {
		if page.IsContentPage() {
			page.AppendTagsToContent()
			page.Save()
		}
	}
}

// buildIndexPage creates the main index.md page that is the root of the site
func buildIndexPage(pageSet []*pages.Page, tagMap *pages.TagMap) {
	src.Info(statusIdxBuild)

	content := ""

	// Write the tag list into the top of the index
	tagLinks := []string{}

	for _, tagName := range tagMap.SortedTagNames() {
		tags := tagMap.Get(tagName)
		if len(tags) > 0 {
			tagLinks = append(tagLinks, tags[0].Link())
		}
	}

	content += strings.Join(tagLinks, ", ")
	content += "\n"

	// Write the page list into the middle of the page
	content += pagesToHTMLUnorderedList(pageSet)
	content += "\n"

	// Write the footer content into the bottom of the index
	content += "\n"
	content += src.Footer()

	// And write the file to disk
	tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, true)
	if err != nil {
		src.Info("Failed to get target dir (1)")
		src.Defeat(err)
	}

	filePath := fmt.Sprintf(
		"%s/index.%s",
		tDir,
		pages.FileExtension,
	)

	err = ioutil.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		src.Info("Failed to write file (1)")
		src.Defeat(err)
	}

	src.Progress(filePath)
}

// buildTagPages creates the tag pages, with links to posts tagged with those names
func buildTagPages(pageSet []*pages.Page) *pages.TagMap {
	src.Info(statusTagBuild)

	tagMap := pages.NewTagMap(pageSet)

	var wGroup sync.WaitGroup

	for _, tagName := range tagMap.SortedTagNames() {
		wGroup.Add(1)

		go func(tagName string) {
			defer wGroup.Done()

			content := fmt.Sprintf("## %s\n\n", tagName)

			// Write the page list into the middle of the page
			content += pagesToHTMLUnorderedList(tagMap.PagesFor(tagName))

			// Write the footer content into the bottom of the page
			content += "\n"
			content += src.Footer()

			// And write the file to disk
			tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, true)
			if err != nil {
				src.Info("Failed to get target dir (2)")
				src.Defeat(err)
			}

			filePath := fmt.Sprintf(
				"%s/%s.%s",
				tDir,
				tagName,
				pages.FileExtension,
			)

			err = ioutil.WriteFile(filePath, []byte(content), 0644)
			if err != nil {
				src.Info("Failed to write file (2)")
				src.Defeat(err)
			}

			src.Progress(filePath)
		}(tagName)
	}

	wGroup.Wait()

	return tagMap
}

func createNewPage(title string) {
	tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, true)
	if err != nil {
		src.Info("Failed to get target dir (3)")
		src.Defeat(err)
	}

	page := pages.NewPage(title, tDir)

	err = page.Open(defaultEditor)
	if err != nil {
		src.Info("Failed to open editor")
		src.Defeat(err)
	}

	// Write the page path to the console. This makes it easy to know which file we just created
	src.Info(page.FilePath)
}

// determineCommitMessage figures out which commit message to save the repo with
// The order of precedence is:
//	* message passed in via the -s flag
//	* message defined in config.yml for the commitMessage key
//	* message as a hard-coded constant, at top, in defaultCommitMsg
// Example:
//  > til -t b -s this is message
func determineCommitMessage(cfg *config.Config, args []string) string {
	if flag.NArg() == 0 {
		return cfg.UString("commitMessage", defaultCommitMsg)
	}

	msgOffset := len(args) - flag.NArg()
	msg := strings.Join(args[msgOffset:], " ")

	return msg
}

// listTargetDirectories writes the list of target directories in the configuration
// out to the terminal
func listTargetDirectories(cfg *config.Config) {
	dirMap, err := cfg.Map("targetDirectories")
	if err != nil {
		src.Info("Failed to map target directories")
		src.Defeat(err)
	}

	for key, dir := range dirMap {
		src.Info(fmt.Sprintf("%6s\t%s\n", key, dir.(string)))
	}
}

// loadPages reads the page files from disk (in reverse chronological order) and
// creates Page instances from them
func loadPages() []*pages.Page {
	pageSet := []*pages.Page{}

	tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, true)
	if err != nil {
		src.Info("Failed to get target dir (4)")
		src.Defeat(err)
	}

	filePaths, _ := filepath.Glob(
		fmt.Sprintf(
			"%s/*.%s",
			tDir,
			pages.FileExtension,
		),
	)

	for i := len(filePaths) - 1; i >= 0; i-- {
		page := pages.PageFromFilePath(filePaths[i])
		pageSet = append(pageSet, page)
	}

	return pageSet
}

// // open tll the OS to open the newly-created page in the editor (as specified in the config)
// // If there's no editor explicitly defined by the user, tell the OS to try and open it
// func open(page *src.Page) error {
// 	editor := src.GlobalConfig.UString("editor", defaultEditor)
// 	if editor == "" {
// 		editor = defaultEditor
// 	}

// 	cmd := exec.Command(editor, page.FilePath)
// 	err := cmd.Run()

// 	return err
// }

// pagesToHTMLUnorderedList creates the unordered list of page links that appear
// on the index and tag pages
func pagesToHTMLUnorderedList(pageSet []*pages.Page) string {
	content := ""
	prevPage := &pages.Page{}

	for _, page := range pageSet {
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
	src.Info(statusRepoPush)

	tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, false)
	if err != nil {
		src.Info("Failed to get target dir (5)")
		src.Defeat(err)
	}

	r, err := git.PlainOpen(tDir)
	if err != nil {
		src.Info("Failed to plain open")
		src.Defeat(err)
	}

	err = r.Push(&git.PushOptions{})
	if err != nil {
		src.Info("Failed to git push")
		src.Defeat(err)
	}
}

// https://github.com/go-git/go-git/blob/master/_examples/commit/main.go
func save(commitMsg string) {
	src.Info(statusRepoSave)

	tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, false)
	if err != nil {
		src.Defeat(err)
	}

	r, err := git.PlainOpen(tDir)
	if err != nil {
		src.Defeat(err)
	}

	w, err := r.Worktree()
	if err != nil {
		src.Defeat(err)
	}

	_, err = w.Add(".")
	if err != nil {
		src.Defeat(err)
	}

	defaultCommitEmail, err2 := src.GlobalConfig.String("committerEmail")
	defaultCommitName, err3 := src.GlobalConfig.String("committerName")
	if err2 != nil || err3 != nil {
		src.Defeat(errors.New(errConfigValueRead))
	}

	commit, err := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  defaultCommitName,
			Email: defaultCommitEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		src.Defeat(err)
	}

	obj, err := r.CommitObject(commit)
	if err != nil {
		src.Defeat(err)
	}

	src.Info(fmt.Sprintf("committed with '%s' (%.7s)", obj.Message, obj.Hash.String()))
}
