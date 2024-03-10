package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/prxg22/git-drive/internal/handlers"
	"github.com/prxg22/git-drive/internal/services"
	"github.com/prxg22/git-drive/pkg/git"
	"github.com/prxg22/git-drive/pkg/spaserver"
)

func main() {
	var _port, _privateKey, _pass, _fileServerPath, _owner, _repo, _remote, _path string

	// get config from flags
	flag.StringVar(&_port, "port", ":8080", "server port to listen. default :8080")
	flag.StringVar(&_privateKey, "key", "", "ssh private key path")
	flag.StringVar(&_pass, "pwd", "", "ssh private key password. optional")
	flag.StringVar(&_fileServerPath, "static", "./app/build/client", "path in whichthe static files are located. default \"./public\"")
	flag.StringVar(&_owner, "owner", "", "repo's owner")
	flag.StringVar(&_repo, "repo", "", "repo's name")
	flag.StringVar(&_remote, "remote", "origin", "repo's remote name")
	flag.StringVar(&_path, "path", "/"+_repo, "local path in which repo will be cloned")
	flag.Parse()

	if _privateKey == "" || _owner == "" || _repo == "" {
		log.Fatalf("missing config: path (%v), owner (%v), repo (%v)", _privateKey, _owner, _repo)
	}

	// initiate git storage
	auth, err := ssh.NewPublicKeysFromFile("git", _privateKey, _pass)
	if err != nil {
		log.Fatal(fmt.Errorf("failed getting keys on path \"%v\": \n%w", _privateKey, err))
	}

	gp := git.NewGitProcessor(_owner, _repo, _remote, _path, auth)
	gs := git.NewGitStorage(gp)
	gds := &services.Service{Storage: gs}

	// initiate routes and server
	routes := make(spaserver.Routes)

	handler := handlers.DirHandler{Service: gds}
	routes["OPTIONS /{dir...}"] = handlers.Options
	routes["GET /dir/{dir...}"] = handler.ReadDir
	routes["GET /dir"] = handler.ReadDir
	routes["DELETE /{path...}"] = handler.Remove

	s := spaserver.NewSPAServer(&routes, "/_api", _fileServerPath)
	log.Printf("listening on port %v\n", _port)
	s.Listen(_port)
}
