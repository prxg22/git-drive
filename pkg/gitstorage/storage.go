package gitstorage

import (
	"fmt"
	"io/fs"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type GithubStorage interface {
	ReadDir(path string) ([]fs.FileInfo, error)
	Remove(path string) error
}

type Storage struct {
	url    string
	remote string
	auth   transport.AuthMethod
	fs     billy.Filesystem
	ms     *memory.Storage
}

type es = *struct{}

var PULL_ACCEPTED_ERRORS = map[string]es{"already up-to-date": nil}

func NewGitStorage(owner, repo, remote string, auth transport.AuthMethod) *Storage {
	var url string

	switch auth.(type) {
	case *ssh.PublicKeys:
		url = fmt.Sprintf("git@github.com:%v/%v.git", owner, repo)
	default:
		url = fmt.Sprintf("https://github.com/%v/%v", owner, repo)
	}

	gs := &Storage{url, remote, auth, memfs.New(), memory.NewStorage()}

	gs.clone()
	return gs
}

func (gs *Storage) ReadDir(path string) ([]fs.FileInfo, error) {
	if _, err := gs.pull(); err != nil {
		return nil, err
	}

	if files, err := gs.fs.ReadDir(path); err == nil {
		return files, nil
	} else {
		return nil, fmt.Errorf("failed to read directory \"%v\": %w", path, err)
	}
}

func (gs *Storage) Remove(path string) error {
	r, err := gs.open()

	if err != nil {
		return err
	}

	w, err := r.Worktree()

	if err != nil {
		return err
	}

	w.Remove(path)

	return gs.commit(".", fmt.Sprintf("removing %v 🗑️", path))
}

func (gs *Storage) open() (*git.Repository, error) {
	r, err := git.Open(gs.ms, gs.fs)

	if err != nil {
		return gs.clone()
	}

	return r, nil
}

func (gs *Storage) pull() (*git.Repository, error) {
	r, err := gs.open()

	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	opts := &git.PullOptions{
		RemoteName: gs.remote,
		Auth:       gs.auth,
	}

	pullErr := w.Pull(opts)
	if pullErr != nil {
		if _, acceptable := PULL_ACCEPTED_ERRORS[pullErr.Error()]; !acceptable {
			return nil, fmt.Errorf("failed to pull working tree: %w", err)
		}
	}

	return r, nil
}

func (gs *Storage) clone() (*git.Repository, error) {
	r, err := git.Clone(gs.ms, gs.fs, &git.CloneOptions{
		URL:        gs.url,
		Auth:       gs.auth,
		RemoteName: gs.remote,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return r, nil
}

func (gs *Storage) commit(path, message string) error {
	r, err := gs.open()

	if err != nil {
		return err
	}

	w, err := r.Worktree()

	if err != nil {
		return err
	}

	if _, err = w.Add(path); err != nil {
		return err
	}

	if _, err = w.Commit(message, &git.CommitOptions{}); err != nil {
		return err
	}

	return r.Push(&git.PushOptions{
		Auth: gs.auth,
	})
}
