package services

import (
	"fmt"
	"strings"

	"github.com/prxg22/git-drive/pkg/git"
)

type GitDriveService interface {
	ReadDir(path string) ([]FileInfo, error)
	Remove(path string) (*Operation, error)
	ListeOperation(id int64) (chan *Operation, error)
}

type Service struct {
	GFS *git.GitFileSystem
	ops map[int64]*Operation
}

type FileInfo struct {
	Name string `json:"name"`
	// size in Mb
	Size  float64 `json:"size"`
	IsDir bool    `json:"isDir"`
}

type Operation struct {
	Id       int64  `json:"id"`
	Op       byte   `json:"op"`
	Progress uint32 `json:"progress"`
	Status   string `json:"status"`
	Data     string `json:"data"`
}

func NewGitDriveService(gfs *git.GitFileSystem) *Service {
	return &Service{
		gfs,
		make(map[int64]*Operation),
	}
}

func (gds *Service) ReadDir(path string) ([]FileInfo, error) {
	if f, err := gds.GFS.ReadDir(strings.TrimSpace(path)); err == nil {
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
	if id, err := gds.GFS.Remove(path); err == nil {
		op := &Operation{
			id,
			'r',
			0,
			"pending",
			"",
		}

		gds.ops[id] = op
		return op, nil
	} else {
		return nil, err
	}

}

func (gds *Service) ListeOperation(id int64) (chan *Operation, error) {
	op := gds.ops[id]

	if op == nil {
		return nil, fmt.Errorf("Operation with id %d not found", id)
	}

	out := make(chan *Operation)

	go func() {
		defer close(out)
		for p := range gds.GFS.Processor.ListenOperation(id) {
			op.Progress = p.Progress
			op.Status = p.Status

			out <- op
		}
	}()

	return out, nil
}
