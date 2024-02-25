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
}

type Storage struct {
	url  string
	auth transport.AuthMethod
	fs   billy.Filesystem
	ms   *memory.Storage
}

func NewGitStorage(owner, repo string, auth transport.AuthMethod) *Storage {
	var url string

	switch auth.(type) {
	case *ssh.PublicKeys:
		url = fmt.Sprintf("git@github.com:%v/%v.git", owner, repo)
	default:
		url = fmt.Sprintf("https://github.com/%v/%v", owner, repo)
	}

	gs := &Storage{url, auth, memfs.New(), memory.NewStorage()}

	gs.clone()
	return gs
}

func (gs *Storage) ReadDir(path string) ([]fs.FileInfo, error) {
	if _, err := gs.open(); err != nil {
		return nil, err
	}

	if files, err := gs.fs.ReadDir(path); err == nil {
		return files, nil
	} else {
		return nil, fmt.Errorf("failed to read directory \"%v\": %w", path, err)
	}
}

func (gs *Storage) open() (*git.Repository, error) {
	r, err := git.Open(gs.ms, gs.fs)

	if err != nil {
		return gs.clone()
	}
	return r, nil
}

func (gs *Storage) clone() (*git.Repository, error) {
	r, err := git.Clone(gs.ms, gs.fs, &git.CloneOptions{
		URL:  gs.url,
		Auth: gs.auth,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return r, nil
}
