package services

import "github.com/prxg22/git-drive/pkg/gitstorage"

type GitDriveService interface {
	ReadDir(path string) ([]FileInfo, error)
}

type Service struct {
	gitstorage.GithubStorage
}

type FileInfo struct {
	Name string `json:"name"`
}

func (gds *Service) ReadDir(path string) ([]FileInfo, error) {
	if f, err := gds.GithubStorage.ReadDir(path); err == nil {
		files := make([]FileInfo, len(f))

		for i, file := range f {
			files[i] = FileInfo{Name: file.Name()}
		}

		return files, nil
	} else {
		return nil, err
	}

}
