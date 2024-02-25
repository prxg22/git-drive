package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/prxg22/git-drive/internal/services"
)

type DirHandler struct {
	S services.GitDriveService
}

func (dh *DirHandler) ReadDir(w http.ResponseWriter, r *http.Request) {
	dir := r.PathValue("dir")

	files, serviceErr := dh.S.ReadDir(dir)

	w.Header().Add("Access-Control-Allow-Origin", "*")

	if serviceErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(serviceErr.Error()))
	}

	if res, parseErr := json.Marshal(files); serviceErr == nil {
		w.WriteHeader(200)
		w.Header().Add("Content-Type", "application/json")
		w.Write(res)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(parseErr.Error()))
	}

}
