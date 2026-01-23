package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/olebedev/config"
)

// Logger defines a minimal logging interface for git operations
type Logger interface {
	Info(msg string)
}

// Repository provides git operations for a repository
type Repository interface {
	Commit(msg string) error
	Push() error
}

// GitRepository implements Repository using go-git
type GitRepository struct {
	repo   *git.Repository
	config *config.Config
	logger Logger
}

// NewRepository opens a git repository at the given directory
func NewRepository(dir string, cfg *config.Config, logger Logger) (*GitRepository, error) {
	r, err := git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	return &GitRepository{
		repo:   r,
		config: cfg,
		logger: logger,
	}, nil
}

// Commit commits all changes in the repository with the given message
func (r *GitRepository) Commit(msg string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = w.Add(".")
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	commitEmail, err := r.config.String("committerEmail")
	if err != nil {
		return fmt.Errorf("could not read committerEmail: %w", err)
	}

	commitName, err := r.config.String("committerName")
	if err != nil {
		return fmt.Errorf("could not read committerName: %w", err)
	}

	commit, err := w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  commitName,
			Email: commitEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	obj, err := r.repo.CommitObject(commit)
	if err != nil {
		return fmt.Errorf("failed to get commit object: %w", err)
	}

	// Log commit info if logger is available
	if r.logger != nil {
		r.logger.Info(fmt.Sprintf("committed with '%s' (%.7s)", obj.Message, obj.Hash.String()))
	}

	return nil
}

// Push pushes changes to the remote repository
func (r *GitRepository) Push() error {
	err := r.repo.Push(&git.PushOptions{})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}
