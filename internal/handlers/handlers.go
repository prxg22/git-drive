package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"path"

	"github.com/prxg22/git-drive/internal/services"
)

type DirHandler struct {
	Service services.GitDriveService
}

func Options(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "*")
	w.WriteHeader(200)
	w.Write([]byte(""))
}

func (dh *DirHandler) ReadDir(w http.ResponseWriter, r *http.Request) {
	dir := r.PathValue("dir")

	files, err := dh.Service.ReadDir(dir)

	w.Header().Add("Access-Control-Allow-Origin", "*")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte(err.Error()))
		return
	}

	if res, err := json.Marshal(files); err == nil {
		w.WriteHeader(200)
		w.Header().Add("Content-Type", "application/json")
		w.Write(res)
	} else {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(err.Error()))
	}
}

func (dh *DirHandler) Remove(w http.ResponseWriter, r *http.Request) {
	path := path.Clean(r.PathValue("path"))
	w.Header().Add("Access-Control-Allow-Origin", "*")

	op, err := dh.Service.Remove(path)

	if err != nil {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if res, err := json.Marshal(op); err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(res)
	} else {
		log.Println(err)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
