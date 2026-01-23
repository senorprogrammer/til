package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/senorprogrammer/til/pages"
	"github.com/senorprogrammer/til/src"
	"github.com/senorprogrammer/til/src/builder"
	"github.com/senorprogrammer/til/src/commands"
	"github.com/senorprogrammer/til/src/git"
	"github.com/senorprogrammer/til/src/loader"
	"github.com/senorprogrammer/til/src/logging"
	"github.com/senorprogrammer/til/src/utils"
)

const (
	defaultCommitMsg = "build, save, push"
	defaultEditor    = "open"
	errNoTitle       = "title must not be blank"
	statusDone       = "done"
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

	logger := logging.NewStdLogger()

	/* Flaghandling */
	/* I personally think "flag handling" should be spelled flag-handling
	   but precedence has been set and we will defer to it.
	   According to wiktionary.org, "stick-handling" is correctly spelled
	   "stickhandling", so here we are, abomination enshrined */

	if listFlag {
		cmd := commands.NewListCommand(src.GlobalConfig, logger)
		if err := cmd.Execute(); err != nil {
			logger.Defeat(err)
		}
		logger.Victory(statusDone)
	}

	if buildFlag {
		tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, true)
		if err != nil {
			logger.Defeat(err)
		}

		loader := loader.NewPageLoader(tDir, pages.FileExtension)
		writer := &builder.OSFileWriter{}
		builder := builder.NewSiteBuilder(tDir, writer, logger)

		cmd := commands.NewBuildCommand(builder, loader, logger)
		if err := cmd.Execute(); err != nil {
			logger.Defeat(err)
		}
		logger.Victory(statusDone)
	}

	if saveFlag {
		// For save, we need the repo root (without docs) for git operations
		repoDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, false)
		if err != nil {
			logger.Defeat(err)
		}

		// But we need the docs directory for building pages
		docsDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, true)
		if err != nil {
			logger.Defeat(err)
		}

		repo, err := git.NewRepository(repoDir, src.GlobalConfig, logger)
		if err != nil {
			logger.Defeat(err)
		}

		loader := loader.NewPageLoader(docsDir, pages.FileExtension)
		writer := &builder.OSFileWriter{}
		builder := builder.NewSiteBuilder(docsDir, writer, logger)
		resolver := commands.NewCommitMessageResolver(src.GlobalConfig, defaultCommitMsg)

		cmd := commands.NewSaveCommand(builder, loader, repo, resolver, os.Args, logger)
		if err := cmd.Execute(); err != nil {
			logger.Defeat(err)
		}
		logger.Victory(statusDone)
	}

	src.BuildTargetDirectory()

	/* Page creation */

	title := utils.ParseTitle(targetDirFlag, os.Args)
	if title == "" {
		// Every non-dash argument is considered a part of the title. If there are no arguments, we have no title
		// Can't have a page without a title
		logger.Defeat(errors.New(errNoTitle))
	}

	tDir, err := src.GetTargetDir(src.GlobalConfig, targetDirFlag, true)
	if err != nil {
		logger.Defeat(err)
	}

	cmd := commands.NewCreateCommand(tDir, title, defaultEditor, logger)
	if err := cmd.Execute(); err != nil {
		logger.Defeat(err)
	}

	logger.Victory(statusDone)
}
