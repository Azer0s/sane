package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

// Repo sane repo
type Repo struct {
	User   string `json:"user"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Tag    string `json:"tag"`
}

// SaneConfig config for sane
type SaneConfig struct {
	Repos []Repo `json:"repos"`
}

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	args := os.Args[1:]

	home, err := homedir.Dir()

	if err != nil {
		log.Fatal(err)
	}

	repoFile := path.Join(home, "./.sane/repos.json")
	repos, err := os.Open(repoFile)

	if err != nil {
		os.Mkdir(path.Join(home, "./.sane/"), os.ModePerm)

		f, err := os.Create(repoFile)

		if err != nil {
			log.Fatal(err)
		}

		var cfgStruct SaneConfig
		cfg, err := json.Marshal(&cfgStruct)

		_, err = f.Write(cfg)

		if err != nil {
			log.Fatal(err)
		}

		f.Close()

		repos, _ = os.Open(repoFile)
	}

	b, err := ioutil.ReadAll(repos)

	if err != nil {
		fmt.Println("File not found!")
	} else {
		fmt.Println(b)
	}

	command := args[0]

	switch command {
	case "start":
		fmt.Println("Start")

	case "list":

	}

	return 0
}
