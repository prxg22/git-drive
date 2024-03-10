package git

import (
	"fmt"
	"log"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/prxg22/git-drive/pkg/queue"
)

type accepted_errors_set = *struct{}

const QUEUE_MAX_SIZE = 10
const FLUSH_TIMEOUT = 1
const PUSH_TIMEOUT = 5

var PULL_ACCEPTED_ERRORS = map[string]accepted_errors_set{"already up-to-date": nil}

type GitProcessor struct {
	FileSystem billy.Filesystem
	url        string
	remote     string
	auth       transport.AuthMethod
	ms         *memory.Storage
	queue      *queue.Queue[CommitCmd]
}

type CommitCmd struct {
	message string
	paths   []string
}

func NewGitProcessor(owner, repo, remote string, auth transport.AuthMethod) *GitProcessor {
	var url string

	switch auth.(type) {
	case *ssh.PublicKeys:
		url = fmt.Sprintf("git@github.com:%v/%v.git", owner, repo)
	default:
		url = fmt.Sprintf("https://github.com/%v/%v", owner, repo)
	}

	q := queue.NewQueue[CommitCmd](QUEUE_MAX_SIZE)
	gp := &GitProcessor{
		FileSystem: memfs.New(),
		url:        url,
		remote:     remote,
		auth:       auth,
		ms:         memory.NewStorage(),
		queue:      q,
	}

	gp.clone()
	go gp.process()

	return gp
}

func (gp *GitProcessor) Commit(cmd CommitCmd) error {
	if gp.queue.IsFull() {
		cmd, err := gp.queue.Dequeue()
		if err != nil {
			return err
		}

		gp.processCmd(cmd)
	}

	log.Printf("enqueing cmd %v", cmd)
	gp.queue.Enqueue(cmd)
	return nil
}

func (gp *GitProcessor) open() (*git.Repository, error) {
	r, err := git.Open(gp.ms, gp.FileSystem)

	if err != nil {
		return gp.clone()
	}

	return r, nil
}

func (gp *GitProcessor) pull() (*git.Repository, error) {
	r, err := gp.open()

	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	opts := &git.PullOptions{
		RemoteName: gp.remote,
		Auth:       gp.auth,
	}

	pullErr := w.Pull(opts)
	if pullErr != nil {
		if _, acceptable := PULL_ACCEPTED_ERRORS[pullErr.Error()]; !acceptable {
			return nil, fmt.Errorf("failed to pull working tree: %w", err)
		}
	}

	return r, nil
}

func (gp *GitProcessor) clone() (*git.Repository, error) {
	r, err := git.Clone(gp.ms, gp.FileSystem, &git.CloneOptions{
		URL:        gp.url,
		Auth:       gp.auth,
		RemoteName: gp.remote,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return r, nil
}

func (gp *GitProcessor) add(paths []string) error {
	r, err := gp.open()

	if err != nil {
		return err
	}

	w, err := r.Worktree()

	if err != nil {
		return err
	}

	for _, path := range paths {
		if _, err = w.Add(path); err != nil {
			return err
		}
	}

	return nil
}

func (gp *GitProcessor) commit(message string) error {
	r, err := gp.open()

	if err != nil {
		return err
	}

	w, err := r.Worktree()

	if err != nil {
		return err
	}

	if _, err = w.Commit(message, &git.CommitOptions{}); err != nil {
		return err
	}

	return nil
}

func (gp *GitProcessor) push() error {
	log.Printf("Pushing... | len: %d", gp.queue.Length())
	if r, err := gp.open(); err == nil {
		return r.Push(&git.PushOptions{
			RemoteName: gp.remote,
			Auth:       gp.auth,
		})
	} else {
		return err
	}
}

func (gp *GitProcessor) processCmd(cmd CommitCmd) error {
	log.Printf("Processing cmd %v | len: %d", cmd, gp.queue.Length())
	if err := gp.add(cmd.paths); err != nil {
		return err
	}

	if err := gp.commit(cmd.message); err != nil {
		return err
	}

	return nil
}

func (gp *GitProcessor) flush() error {
	if gp.queue.Length() == 0 {
		return fmt.Errorf("empty queue")
	}

	log.Printf("flushing | q.length: %d", gp.queue.Length())

	return gp.queue.Flush(func(cmd CommitCmd, _ int) error {
		return gp.processCmd(cmd)
	})
}

func (gp *GitProcessor) process() {
	flushDuration := FLUSH_TIMEOUT * time.Second
	pushDuration := PUSH_TIMEOUT * time.Second
	flushTimer := time.NewTimer(flushDuration)
	pushTimer := time.NewTimer(pushDuration)

	for {
		select {
		case <-pushTimer.C:
			pushTimer.Reset(pushDuration)
			log.Println("processor pushing...")

			if err := gp.push(); err != nil {
				log.Println(err)
			}
		case <-flushTimer.C:
			flushTimer.Reset(flushDuration)
			log.Println("processor flushing...")

			if err := gp.flush(); err != nil {
				log.Println(err)
				continue
			}
		default:
			log.Println("processor pulling...")
			gp.pull()
		}
	}
}
