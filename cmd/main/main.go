package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/prxg22/go-drive/pkg/gitstorage"
)

func main() {
	var path, pass, owner, repo string
	flag.StringVar(&path, "pk", "", "ssh private key path")
	flag.StringVar(&pass, "p", "", "ssh private key password")
	flag.StringVar(&owner, "o", "", "repo's owner")
	flag.StringVar(&repo, "r", "", "repo's name")
	flag.Parse()

	if path == "" || owner == "" || repo == "" {
		log.Fatalf("missing config: path (%v), owner (%v), repo (%v)", path, owner, repo)
	}
	keys, err := ssh.NewPublicKeysFromFile("git", path, pass)
	log.Println(path, pass, owner, repo)
	if err != nil {
		log.Fatal(fmt.Errorf("failed getting keys on path \"%v\": \n%w", path, err))
	}

	_, err = gitstorage.NewGitStorage(owner, repo, keys)

	if err != nil {
		log.Fatal(err)
	}
}
