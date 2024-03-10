package git

import (
	"fmt"
	"io/fs"
	"os"
	"path"
)

// Storage is an interface that defines the methods for interacting with the Git storage.
type Storage interface {
	ReadDir(path string) ([]fs.FileInfo, error)
	Remove(path string) error
}

// GitStorage represents the Git storage.
type GitStorage struct {
	Path      string        // Path is the root path of the Git storage.
	Processor *GitProcessor // Processor is the Git processor associated with the storage.
}

// NewGitStorage creates a new instance of GitStorage.
// It takes a pointer to a GitProcessor and returns a pointer to a GitStorage.
func NewGitStorage(processor *GitProcessor) *GitStorage {
	return &GitStorage{Path: path.Clean(processor.Path), Processor: processor}
}

// ReadDir reads the contents of a directory specified by the given path.
// It returns a slice of fs.FileInfo representing the files and directories in the directory.
// If the path is "/", it reads the root directory.
// The function excludes the ".git" directory from the result.
func (gs *GitStorage) ReadDir(p string) ([]fs.FileInfo, error) {
	if p == "/" {
		p = ""
	}

	path := path.Join(gs.Path, p)
	dirs, err := os.ReadDir(path)

	if err != nil {
		return nil, fmt.Errorf("failed to read directory \"%v\": %w", path, err)
	}

	size := len(dirs)

	if p == "" {
		size -= 1
	}

	var infos = make([]fs.FileInfo, size)

	i := 0
	for _, dir := range dirs {
		if dir.Name() == ".git" {
			continue
		}

		infos[i], err = dir.Info()

		if err != nil {
			return nil, err
		}

		i++
	}

	return infos, nil
}

// Remove removes a file or directory from the Git storage.
// It returns the commit operation ID and any error encountered.
func (gs *GitStorage) Remove(p string) (int64, error) {
	gp := gs.Processor

	info, err := os.Stat(path.Join(gs.Path, p))

	if err != nil {
		return -1, err
	}

	if info.IsDir() {
		err = os.RemoveAll(path.Join(gs.Path, p))
	} else {
		err = os.Remove(path.Join(gs.Path, p))
	}

	if err != nil {
		return -1, err
	}

	id := gp.Commit(
		"rm: "+p,
		[]string{p},
	)

	return id, nil
}
