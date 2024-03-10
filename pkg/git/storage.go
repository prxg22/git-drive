package git

import (
	"fmt"
	"io/fs"
)

type Storage interface {
	ReadDir(path string) ([]fs.FileInfo, error)
	Remove(path string) error
}

type GitStorage struct {
	Processor *GitProcessor
}

func (gs *GitStorage) ReadDir(path string) ([]fs.FileInfo, error) {
	gp := gs.Processor

	if files, err := gp.FileSystem.ReadDir(path); err == nil {
		return files, nil
	} else {
		return nil, fmt.Errorf("failed to read directory \"%v\": %w", path, err)
	}
}

func (gs *GitStorage) Remove(path string) error {
	gp := gs.Processor

	gp.FileSystem.Remove(path)

	gp.Commit(CommitCmd{
		message: "rm: " + path,
		paths:   []string{path},
	})

	return nil
}
