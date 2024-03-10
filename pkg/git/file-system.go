package git

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
)

// Storage is an interface that defines the methods for interacting with the Git storage.
type Storage interface {
	ReadDir(path string) ([]fs.FileInfo, error)
	Remove(path string) error
}

// GitFileSystem represents the Git storage.
type GitFileSystem struct {
	Path      string     // Path is the root path of the Git storage.
	Processor *GitClient // Processor is the Git processor associated with the storage.
}

// NewGitFileSystem creates a new instance of GitStorage.
// It takes a pointer to a GitProcessor and returns a pointer to a GitStorage.
func NewGitFileSystem(processor *GitClient) *GitFileSystem {
	return &GitFileSystem{Path: path.Clean(processor.Path), Processor: processor}
}

// ReadDir reads the contents of a directory specified by the given path.
// It returns a slice of fs.FileInfo representing the files and directories in the directory.
// If the path is "/", it reads the root directory.
// The function excludes the ".git" directory from the result.
func (gfs *GitFileSystem) ReadDir(p string) ([]fs.FileInfo, error) {
	if p == "/" {
		p = ""
	}

	path := path.Join(gfs.Path, p)
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

func (gfs *GitFileSystem) removeRecursively(p string) ([]string, error) {
	log.Printf("trying to remove path \"%s\"", p)
	info, err := os.Stat(path.Join(gfs.Path, p))

	if err != nil {
		return nil, err
	}

	paths := []string{}

	if info.IsDir() {

		if dirs, err := os.ReadDir(path.Join(gfs.Path, p)); err == nil {
			for _, dir := range dirs {
				if dirPaths, err := gfs.removeRecursively(path.Join(p, dir.Name())); err == nil {
					paths = append(paths, dirPaths...)
				} else {
					return nil, err
				}
			}
		} else {
			return nil, err
		}

	} else {
		paths = append(paths, p)
	}

	if err := os.Remove(path.Join(gfs.Path, p)); err != nil {
		return nil, err
	}

	log.Printf("removed path \"%s\" | paths: %v", p, paths)

	return paths, nil
}

// Remove removes a file or directory from the Git storage.
// It returns the commit operation ID and any error encountered.
func (gfs *GitFileSystem) Remove(p string) (int64, error) {
	gp := gfs.Processor

	paths, err := gfs.removeRecursively(p)

	if err != nil {
		return -1, err
	}

	id := gp.Commit(
		"rm: "+strings.Join(paths, " | "),
		paths,
	)

	return id, nil
}
