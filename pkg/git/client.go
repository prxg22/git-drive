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
const PUSH_TIMEOUT = 5

var PULL_ACCEPTED_ERRORS = map[string]accepted_errors_set{"already up-to-date": nil}

// GitClient represents a processor for Git operations.
type GitClient struct {
	Path   string                 // Path to the Git repository.
	url    string                 // URL of the remote repository.
	remote string                 // Name of the remote repository.
	auth   transport.AuthMethod   // Authentication method for accessing the remote repository.
	repo   *git.Repository        // Git repository object.
	queue  *queue.Queue[*command] // Queue of commit commands.
	cmds   chan *command          // Channel to receive commit commands.
	out    map[int64]chan *Operation
	ops    map[int64]*Operation
}

type command struct {
	id      int64
	message string
	paths   []string
}

type Operation struct {
	Stage    string
	Status   string
	Progress uint32
	Data     string
}

// NewGitClient creates a new instance of GitProcessor with the specified parameters.
// It initializes the GitProcessor struct and starts a goroutine to process the commands.
// The owner and repo parameters specify the GitHub repository to work with.
// The remote parameter specifies the remote name of the repository.
// The basePath parameter specifies the base path of the local repository.
// The auth parameter specifies the authentication method to use when interacting with the repository.
// It returns a pointer to the created GitProcessor instance.
func NewGitClient(owner, repo, remote, basePath string, auth transport.AuthMethod) *GitClient {
	var url string

	switch auth.(type) {
	case *ssh.PublicKeys:
		url = fmt.Sprintf("git@github.com:%v/%v.git", owner, repo)
	default:
		url = fmt.Sprintf("https://github.com/%v/%v", owner, repo)
	}

	q := queue.NewQueue[*command](QUEUE_MAX_SIZE)
	r, err := open(basePath, url, remote, auth)

	if err != nil {
		log.Panic(err.Error())
	}

	cmds := make(chan *command, QUEUE_MAX_SIZE)
	out := make(map[int64]chan *Operation)
	ops := make(map[int64]*Operation)

	gc := &GitClient{
		Path:   path.Clean(basePath),
		auth:   auth,
		cmds:   cmds,
		queue:  q,
		repo:   r,
		out:    out,
		ops:    ops,
		remote: remote,
		url:    url,
	}

	go gc.process()

	return gc
}

// Commit adds and commits changes asynchronously. It takes a commit message and a list of paths to files that have been changed.
// The function creates a commit command and sends it to the command channel for processing.
// It returns the ID of the commit operation for tracking purposes.
func (gc *GitClient) Commit(message string, paths []string) (int64, error) {
	cmd := &command{
		time.Now().UnixMilli(),
		message,
		paths,
	}

	go func() {
		gc.cmds <- cmd
	}()

	out := make(chan *Operation, QUEUE_MAX_SIZE)
	gc.out[cmd.id] = out
	gc.ops[cmd.id] = &Operation{
		Stage: "pending",
	}

	err := gc.updateOpStage(cmd.id, "queue", 0)

	return cmd.id, err
}

func (gc *GitClient) ListenOperation(id int64) chan *Operation {
	return gc.out[id]
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

func (gc *GitClient) pull() error {
	w, err := gc.repo.Worktree()
	if err != nil {
		return err
	}

	opts := &git.PullOptions{
		RemoteName: gc.remote,
		Auth:       gc.auth,
	}

	dff := w.Pull(opts)
	if dff != nil {
		if _, acceptable := PULL_ACCEPTED_ERRORS[dff.Error()]; !acceptable {
			return fmt.Errorf("failed to pull working tree: %w", err)
		}
	}

	return nil
}

func (gc *GitClient) add(paths []string) error {
	w, err := gc.repo.Worktree()

	if err != nil {
		return err
	}

	for _, p := range paths {
		if err != nil {
			return fmt.Errorf("error getting path %s stats: %w", p, err)
		}

		_, err = w.Add(p)
		// log.Printf("added path \"%s\"", p)

		if err != nil {
			return fmt.Errorf("error adding path %s: %w", p, err)
		}
	}

	return nil
}

func (gc *GitClient) commit(message string) error {
	w, err := gc.repo.Worktree()

	if err != nil {
		return err
	}

	if _, err = w.Commit(message, &git.CommitOptions{}); err != nil {
		return err
	}

	return nil
}

func (gc *GitClient) push() error {
	return gc.repo.Push(&git.PushOptions{
		RemoteName: gc.remote,
		Auth:       gc.auth,
	})
}

func (gc *GitClient) updateOp(id int64, stage string, pgrss uint32, status string, data string) error {
	p, exists := gc.ops[id]

	if !exists {
		return fmt.Errorf("progress %d doesn't exist in progress map", id)
	}

	out, exists := gc.out[id]

	if !exists {
		return fmt.Errorf("out channel %d doesn't exist in progress map", id)
	}
	p.Progress = pgrss
	p.Stage = stage
	p.Status = status
	p.Data = data
	out <- p
	return nil
}

func (gc *GitClient) updateOpStatus(id int64, status string, p int32, data string) error {
	if op, ok := gc.ops[id]; ok {
		prgss := uint32(p)

		if p < 0 {
			prgss = op.Progress
		}

		return gc.updateOp(id, op.Stage, prgss, status, data)
	} else {
		return fmt.Errorf("op %d not found", id)
	}
}

func (gc *GitClient) updateOpStage(id int64, stage string, p uint32) error {
	if op, ok := gc.ops[id]; ok {
		return gc.updateOp(id, stage, p, "pending", op.Data)
	} else {
		return fmt.Errorf("op %d not found", id)
	}
}

// processCmd processes the commit command.
// It adds the specified paths to the Git repository and commits the changes with the given message.
// Returns an error if any operation fails.
func (gc *GitClient) processCmd(cmd *command) error {
	paths := cmd.paths
	if err := gc.add(paths); err != nil {
		gc.updateOpStatus(cmd.id, "failed", -1, err.Error())
		return err
	}
	gc.updateOpStage(cmd.id, "add", 33)

	if err := gc.commit(cmd.message); err != nil {
		gc.updateOpStatus(cmd.id, "failed", -1, err.Error())
		return err
	}
	gc.updateOpStage(cmd.id, "commit", 66)

	return nil
}

func (gc *GitClient) pushCmds() error {
	for cmd := range gc.queue.Iterate() {
		gc.updateOpStage(cmd.id, "push", 69)
		gc.queue.Enqueue(cmd)
	}

	if err := gc.push(); err == nil {
		for cmd := range gc.queue.Iterate() {
			gc.updateOpStatus(cmd.id, "success", 100, "")

			defer delete(gc.out, cmd.id)
			defer delete(gc.ops, cmd.id)
			defer close(gc.out[cmd.id])
		}
	} else if err.Error() != "already up-to-date" {
		for cmd := range gc.queue.Iterate() {
			gc.updateOpStatus(cmd.id, "falied", -1, err.Error())

			defer delete(gc.out, cmd.id)
			defer delete(gc.ops, cmd.id)
			defer close(gc.out[cmd.id])
		}
		return err
	}

	return nil
}

// process is a method of the GitProcessor struct that continuously processes commands from the cmds channel.
// It also handles pushing and pulling changes to and from the remote repository.
// This method runs in an infinite loop until the program is terminated.
func (gc *GitClient) process() {
	pushTimer := time.NewTimer(PUSH_TIMEOUT * time.Second)
	for {
		select {
		case cmd := <-gc.cmds:
			if err := gc.processCmd(cmd); err == nil {
				gc.queue.Enqueue(cmd)
			} else {
				log.Println(fmt.Errorf("error while processor try to process command: %w", err))
			}

		case <-pushTimer.C:
			pushTimer.Reset(PUSH_TIMEOUT * time.Second)
			if err := gc.pushCmds(); err != nil {
				log.Println(fmt.Errorf("error while processor try to push commands: %w", err))
			}

		default:
			if err := gc.pull(); err != nil && err.Error() != "already up-to-date" {
				log.Println(fmt.Errorf("error while processor try to pull: %w", err))
			}
		}
	}
}
