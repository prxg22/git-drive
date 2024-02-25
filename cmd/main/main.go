package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/prxg22/git-drive/internal/handlers"
	"github.com/prxg22/git-drive/internal/services"
	"github.com/prxg22/git-drive/pkg/gitstorage"
	"github.com/prxg22/git-drive/pkg/httpserver"
)

func main() {
	var _port, _privateKey, _pass, _fileServerPath, _owner, _repo string

	// get config from flags
	flag.StringVar(&_port, "p", ":8080", "server port to listen. default :8080")
	flag.StringVar(&_privateKey, "pk", "", "ssh private key path")
	flag.StringVar(&_pass, "ps", "", "ssh private key password. optional")
	flag.StringVar(&_fileServerPath, "fp", "./public", "path where are the static files for the static file. default \"./public\"")
	flag.StringVar(&_owner, "o", "", "repo's owner")
	flag.StringVar(&_repo, "r", "", "repo's name")
	flag.Parse()

	if _privateKey == "" || _owner == "" || _repo == "" {
		log.Fatalf("missing config: path (%v), owner (%v), repo (%v)", _privateKey, _owner, _repo)
	}

	// initiate git storage
	auth, err := ssh.NewPublicKeysFromFile("git", _privateKey, _pass)
	if err != nil {
		log.Fatal(fmt.Errorf("failed getting keys on path \"%v\": \n%w", _privateKey, err))
	}

	gs := gitstorage.NewGitStorage(_owner, _repo, auth)
	gds := &services.Service{GithubStorage: gs}

	// initiate routes and server
	routes := make(httpserver.Routes)

	handler := handlers.DirHandler{S: gds}
	routes["GET /dir/{dir...}"] = handler.ReadDir
	routes["GET /dir"] = handler.ReadDir

	s := httpserver.NewServer(&routes, "/_api", _fileServerPath)
	log.Printf("listening on port %v\n", _port)
	s.Listen(_port)
}
