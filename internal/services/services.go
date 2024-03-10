package services

import (
	"strings"

	"github.com/prxg22/git-drive/pkg/git"
)

type GitDriveService interface {
	ReadDir(path string) ([]FileInfo, error)
	Remove(path string) (*Operation, error)
}

type Service struct {
	Storage *git.GitStorage
}

type FileInfo struct {
	Name string `json:"name"`
	// size in Mb
	Size  float64 `json:"size"`
	IsDir bool    `json:"isDir"`
}

type Operation struct {
	Id int64 `json:"id"`
	Op byte  `json:"op"`
}

func (gds *Service) ReadDir(path string) ([]FileInfo, error) {
	if f, err := gds.Storage.ReadDir(strings.TrimSpace(path)); err == nil {
		files := make([]FileInfo, len(f))

		for i, file := range f {
			files[i] = FileInfo{Name: file.Name(), IsDir: file.IsDir(), Size: float64(file.Size()) / 1280}
		}

		return files, nil
	} else {
		return nil, err
	}
}

func (gds *Service) Remove(path string) (*Operation, error) {
	if id, err := gds.Storage.Remove(path); err == nil {
		op := &Operation{
			id,
			'r',
		}

		return op, nil
	} else {
		return nil, err
	}

}
