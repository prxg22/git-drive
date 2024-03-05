package services

import "github.com/prxg22/git-drive/pkg/gitstorage"

type GitDriveService interface {
	ReadDir(path string) ([]FileInfo, error)
	Remove(path string) error
}

type Service struct {
	gitstorage.GithubStorage
}

type FileInfo struct {
	Name string `json:"name"`
	// size in Mb
	Size  float64 `json:"size"`
	IsDir bool    `json:"isDir"`
}

func (gds *Service) ReadDir(path string) ([]FileInfo, error) {
	if f, err := gds.GithubStorage.ReadDir(path); err == nil {
		files := make([]FileInfo, len(f))

		for i, file := range f {
			files[i] = FileInfo{Name: file.Name(), IsDir: file.IsDir(), Size: float64(file.Size()) / 1280}
		}

		return files, nil
	} else {
		return nil, err
	}
}

func (gds *Service) Remove(path string) error {
	return gds.GithubStorage.Remove(path)
}
