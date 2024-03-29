package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"

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

func (dh *DirHandler) GetOperations(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	c, err := dh.Service.ListeOperation(id)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// // Create a ticker that ticks every second.
	// ticker := time.NewTicker(1 * time.Second)
	// defer ticker.Stop()
	// // Send events for a limited time (e.g., 10 seconds) for demonstration.
	// endTime := time.Now().Add(10 * time.Second)
	// for {
	// 	select {
	// 	case t := <-ticker.C:
	// 		if time.Now().After(endTime) {
	// 			fmt.Fprintln(w, "event: close\ndata: close")
	// 			if f, ok := w.(http.Flusher); ok {
	// 				f.Flush()
	// 			}
	// 			return // Stop the handler.
	// 		}
	// 		fmt.Fprintf(w, "data: The time is: %v\n\n", t)
	// 		if f, ok := w.(http.Flusher); ok {
	// 			f.Flush()
	// 		}
	// 	case <-r.Context().Done():
	// 		return
	// 	}
	// }

	checkProgress(w, r, c)
}

func checkProgress(w http.ResponseWriter, r *http.Request, c chan *services.Operation) {
	ok := true
	for ok {
		select {
		case op := <-c:
			if op == nil {
				fmt.Fprintf(w, "event: close\n\n")
				fmt.Fprintf(w, "data: close\n\n")
				return
			}
			if res, err := json.Marshal(op); err == nil {
				if op.Status == "failed" {
					fmt.Fprintf(w, "event: error\n\n")
				}
				fmt.Fprintf(w, "data: %s\n\n", res)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			} else {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}
