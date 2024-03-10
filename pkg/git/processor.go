package git

import (
	"fmt"
	"log"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/prxg22/git-drive/pkg/queue"
)

type accepted_errors_set = *struct{}

const QUEUE_MAX_SIZE = 20
const FLUSH_TIMEOUT = 1
const PUSH_TIMEOUT = 5

var PULL_ACCEPTED_ERRORS = map[string]accepted_errors_set{"already up-to-date": nil}

// GitProcessor represents a processor for Git operations.
type GitProcessor struct {
	Path   string                   // Path to the Git repository.
	Out    chan []int64             // Channel to send output data.
	url    string                   // URL of the remote repository.
	remote string                   // Name of the remote repository.
	auth   transport.AuthMethod     // Authentication method for accessing the remote repository.
	repo   *git.Repository          // Git repository object.
	queue  *queue.Queue[*commitCmd] // Queue of commit commands.
	cmds   chan *commitCmd          // Channel to receive commit commands.
}

type commitCmd struct {
	id      int64
	message string
	paths   []string
}

// NewGitProcessor creates a new instance of GitProcessor with the specified parameters.
// It initializes the GitProcessor struct and starts a goroutine to process the commands.
// The owner and repo parameters specify the GitHub repository to work with.
// The remote parameter specifies the remote name of the repository.
// The basePath parameter specifies the base path of the local repository.
// The auth parameter specifies the authentication method to use when interacting with the repository.
// It returns a pointer to the created GitProcessor instance.
func NewGitProcessor(owner, repo, remote, basePath string, auth transport.AuthMethod) *GitProcessor {
	var url string

	switch auth.(type) {
	case *ssh.PublicKeys:
		url = fmt.Sprintf("git@github.com:%v/%v.git", owner, repo)
	default:
		url = fmt.Sprintf("https://github.com/%v/%v", owner, repo)
	}

	q := queue.NewQueue[*commitCmd](QUEUE_MAX_SIZE)
	r, err := open(basePath, url, remote, auth)

	if err != nil {
		log.Panic(err.Error())
	}

	cmds := make(chan *commitCmd, QUEUE_MAX_SIZE)
	out := make(chan []int64, QUEUE_MAX_SIZE)

	gp := &GitProcessor{
		Out:    out,
		Path:   path.Clean(basePath),
		auth:   auth,
		queue:  q,
		repo:   r,
		remote: remote,
		url:    url,
		cmds:   cmds,
	}

	go gp.process()

	return gp
}

// Commit adds and commits changes asynchronously. It takes a commit message and a list of paths to files that have been changed.
// The function creates a commit command and sends it to the command channel for processing.
// It returns the ID of the commit operation for tracking purposes.
func (gp *GitProcessor) Commit(message string, paths []string) int64 {
	cmd := &commitCmd{
		time.Now().UnixMilli(),
		message,
		paths,
	}

	go func() {
		gp.cmds <- cmd
	}()

	return cmd.id
}

func open(p, u, r string, a transport.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainOpen(p)

	if err == git.ErrRepositoryNotExists {
		return clone(p, u, r, a)
	} else if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return repo, nil
}

func clone(p, u, r string, a transport.AuthMethod) (*git.Repository, error) {
	repo, err := git.PlainClone(p, false, &git.CloneOptions{
		URL:        u,
		Auth:       a,
		RemoteName: r,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	return repo, nil
}

func (gp *GitProcessor) pull() error {
	w, err := gp.repo.Worktree()
	if err != nil {
		return err
	}

	opts := &git.PullOptions{
		RemoteName: gp.remote,
		Auth:       gp.auth,
	}

	dff := w.Pull(opts)
	if dff != nil {
		if _, acceptable := PULL_ACCEPTED_ERRORS[dff.Error()]; !acceptable {
			return fmt.Errorf("failed to pull working tree: %w", err)
		}
	}

	return nil
}

func (gp *GitProcessor) add(paths []string) error {
	w, err := gp.repo.Worktree()

	if err != nil {
		return err
	}

	for _, p := range paths {
		if _, err = w.Add(p); err != nil {
			return fmt.Errorf("error adding path %s: %w", path.Join("./", gp.Path, p), err)
		}
	}

	return nil
}

func (gp *GitProcessor) commit(message string) error {
	w, err := gp.repo.Worktree()

	if err != nil {
		return err
	}

	if _, err = w.Commit(message, &git.CommitOptions{}); err != nil {
		return err
	}

	return nil
}

func (gp *GitProcessor) push() error {
	return gp.repo.Push(&git.PushOptions{
		RemoteName: gp.remote,
		Auth:       gp.auth,
	})
}

// processCmd processes the commit command.
// It adds the specified paths to the Git repository and commits the changes with the given message.
// Returns an error if any operation fails.
func (gp *GitProcessor) processCmd(cmd *commitCmd) error {
	paths := cmd.paths

	if err := gp.add(paths); err != nil {
		return err
	}

	if err := gp.commit(cmd.message); err != nil {
		return err
	}

	return nil
}

// process is a method of the GitProcessor struct that continuously processes commands from the cmds channel.
// It also handles pushing and pulling changes to and from the remote repository.
// This method runs in an infinite loop until the program is terminated.
func (gp *GitProcessor) process() {
	pushTimer := time.NewTimer(PUSH_TIMEOUT * time.Second)
	for {
		select {
		case cmd := <-gp.cmds:
			if err := gp.processCmd(cmd); err == nil {
				gp.queue.Enqueue(cmd)
			} else {
				log.Println(fmt.Errorf("error while processor try to process command: %w", err))
			}

		case <-pushTimer.C:
			pushTimer.Reset(PUSH_TIMEOUT * time.Second)
			if err := gp.push(); err == nil {
				ids := make([]int64, gp.queue.Length())
				i := 0
				for cmd := range gp.queue.Iterate() {
					ids[i] = cmd.id
					i++
				}

				gp.Out <- ids
			} else if err.Error() != "already up-to-date" {
				log.Println(fmt.Errorf("error while processor try to push: %w", err))
			}
		default:
			if err := gp.pull(); err != nil && err.Error() != "already up-to-date" {
				log.Println(fmt.Errorf("error while processor try to pull: %w", err))

			}
		}
	}
}
