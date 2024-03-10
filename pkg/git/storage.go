package git

import (
	"fmt"
	"io/fs"
	"os"
	"path"
)

type Storage interface {
	ReadDir(path string) ([]fs.FileInfo, error)
	Remove(path string) error
}

type GitStorage struct {
	Path      string
	Processor *GitProcessor
}

func NewGitStorage(processor *GitProcessor) *GitStorage {
	return &GitStorage{Path: path.Clean(processor.Path), Processor: processor}

}

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
