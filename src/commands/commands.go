package commands

import (
	"flag"
	"fmt"
	"strings"

	"github.com/olebedev/config"
	"github.com/senorprogrammer/til/pages"
	"github.com/senorprogrammer/til/src/builder"
	"github.com/senorprogrammer/til/src/git"
	"github.com/senorprogrammer/til/src/loader"
	"github.com/senorprogrammer/til/src/logging"
)

// CommitMessageResolver determines the commit message to use based on precedence
type CommitMessageResolver struct {
	config     *config.Config
	defaultMsg string
}

// NewCommitMessageResolver creates a new CommitMessageResolver
func NewCommitMessageResolver(cfg *config.Config, defaultMsg string) *CommitMessageResolver {
	return &CommitMessageResolver{
		config:     cfg,
		defaultMsg: defaultMsg,
	}
}

// Resolve determines the commit message based on precedence:
// 1. Message passed in via non-flag arguments
// 2. Message defined in config.yml for the commitMessage key
// 3. Default hard-coded constant
func (r *CommitMessageResolver) Resolve(args []string, nonFlagArgs int) string {
	if nonFlagArgs == 0 {
		return r.config.UString("commitMessage", r.defaultMsg)
	}

	msgOffset := len(args) - nonFlagArgs
	return strings.Join(args[msgOffset:], " ")
}

// Command defines the interface for all commands
type Command interface {
	Execute() error
}

// ListCommand lists target directories
type ListCommand struct {
	config *config.Config
	logger logging.Logger
}

// NewListCommand creates a new ListCommand
func NewListCommand(cfg *config.Config, logger logging.Logger) *ListCommand {
	return &ListCommand{
		config: cfg,
		logger: logger,
	}
}

// Name returns the name of the command
func (c *ListCommand) Name() string {
	return "list"
}

// Execute lists the configured target directories from the configuration
func (c *ListCommand) Execute() error {
	dirMap, err := c.config.Map("targetDirectories")
	if err != nil {
		return err
	}

	for key, dir := range dirMap {
		c.logger.Info(fmt.Sprintf("%6s\t%s", key, dir.(string)))
	}

	return nil
}

// BuildCommand builds index and tag pages
type BuildCommand struct {
	builder *builder.SiteBuilder
	loader  *loader.PageLoader
	logger  logging.Logger
}

// NewBuildCommand creates a new BuildCommand
func NewBuildCommand(b *builder.SiteBuilder, l *loader.PageLoader, logger logging.Logger) *BuildCommand {
	return &BuildCommand{
		builder: b,
		loader:  l,
		logger:  logger,
	}
}

// Name returns the name of the command
func (c *BuildCommand) Name() string {
	return "build"
}

// Execute builds the index and tag pages from loaded page data
func (c *BuildCommand) Execute() error {
	pageSet, err := c.loader.Load()
	if err != nil {
		return err
	}

	tagMap := pages.NewTagMap(pageSet)

	if err := c.builder.BuildTagPages(tagMap); err != nil {
		return err
	}

	return c.builder.BuildIndex(pageSet, tagMap)
}

// SaveCommand builds, commits, and pushes changes
type SaveCommand struct {
	builder    *builder.SiteBuilder
	pageLoader *loader.PageLoader
	repo       git.Repository
	resolver   *CommitMessageResolver
	args       []string
	logger     logging.Logger
}

// NewSaveCommand creates a new SaveCommand
func NewSaveCommand(
	b *builder.SiteBuilder,
	l *loader.PageLoader,
	r git.Repository,
	resolver *CommitMessageResolver,
	args []string,
	logger logging.Logger,
) *SaveCommand {
	return &SaveCommand{
		builder:    b,
		pageLoader: l,
		repo:       r,
		resolver:   resolver,
		args:       args,
		logger:     logger,
	}
}

// Name returns the name of the command
func (c *SaveCommand) Name() string {
	return "save"
}

// Execute builds the index and tag pages, then commits and pushes changes to the git repository
func (c *SaveCommand) Execute() error {
	// Build content
	pageSet, err := c.pageLoader.Load()
	if err != nil {
		return err
	}

	tagMap := pages.NewTagMap(pageSet)

	if err := c.builder.BuildTagPages(tagMap); err != nil {
		return err
	}

	if err := c.builder.BuildIndex(pageSet, tagMap); err != nil {
		return err
	}

	// Commit and push
	commitMsg := c.resolver.Resolve(c.args, flag.NArg())

	c.logger.Info("saving uncommitted files")
	if err := c.repo.Commit(commitMsg); err != nil {
		return err
	}

	c.logger.Info("pushing to remote")
	return c.repo.Push()
}

// CreateCommand creates a new page
type CreateCommand struct {
	targetDir string
	title     string
	editor    string
	logger    logging.Logger
}

// NewCreateCommand creates a new CreateCommand
func NewCreateCommand(targetDir, title, editor string, logger logging.Logger) *CreateCommand {
	return &CreateCommand{
		targetDir: targetDir,
		title:     title,
		editor:    editor,
		logger:    logger,
	}
}

// Name returns the name of the command
func (c *CreateCommand) Name() string {
	return "create"
}

// Execute creates a new page file and opens it in the configured editor
func (c *CreateCommand) Execute() error {
	page := pages.NewPage(c.title, c.targetDir)

	err := page.Open(c.editor)
	if err != nil {
		return err
	}

	c.logger.Info(page.FilePath)
	return nil
}
