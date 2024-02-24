package gitstorage

import (
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type GitStorage struct {
	owner string
	repo  string
	url   string
	keys  *ssh.PublicKeys
}

func NewGitStorage(owner string, repo string, keys *ssh.PublicKeys) (*GitStorage, error) {

	url := fmt.Sprintf("git@github.com:%v/%v.git", owner, repo)

	st := &GitStorage{owner, repo, url, keys}	

	if err := st.setup(); err != nil {
		return nil, err
	}

	return st, nil
}

func (gs *GitStorage) setup() error {
	options := git.CloneOptions{
		URL:  gs.url,
		Auth: gs.keys,
		Progress: os.Stdout,
		SingleBranch: true,
	}

	ms := memory.NewStorage()
	fs := memfs.New()

	log.Println("cloning repo...")
	r, err := git.Clone(ms, fs, &options)

	if err != nil {
		return fmt.Errorf("failed cloning the repo: %w", err)
	}

	log.Println("trying to get HEAD ref...")
	ref, err := r.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	log.Println("logging commits...")
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})

	if err != nil {
		return fmt.Errorf("failed logging: %w", err)
	}

	log.Println("iterating over commits...")
	// ... just iterates over the commits, printing it
	if err = cIter.ForEach(func(c *object.Commit) error {
		log.Println(c)
		return nil
	}); err != nil {
		return fmt.Errorf("failed iterating over commits: %w", err)

	}

	return nil
}
