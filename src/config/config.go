package config

import (
	"../util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

const (
	//APPLY apply constant
	APPLY = "apply"
	//REMOVE remove constant
	REMOVE = "remove"
)

// Repo sane repo
type Repo struct {
	User   string   `json:"user"`
	Name   string   `json:"name"`
	Branch string   `json:"branch"`
	Tag    string   `json:"tag"`
	Topics []string `json:"topics"`
}

// SaneConfig config for sane
type SaneConfig struct {
	Repos   []Repo            `json:"repos"`
	Aliases map[string]string `json:"aliases"`
}

//Contains check if array contains repo
func Contains(arr []Repo, item Repo) bool {
	for _, a := range arr {
		if a.User == item.User && a.Branch == item.Branch && a.Name == item.Name && a.Tag == item.Tag {
			return true
		}
	}
	return false
}

//IndexOf repo in array
func IndexOf(arr []Repo, item Repo) int {
	for i, a := range arr {
		if a.User == item.User && a.Branch == item.Branch && a.Name == item.Name && a.Tag == item.Tag {
			return i
		}
	}
	return -1
}

//Mapkeys get keys by value
func Mapkeys(m map[string]string, value string) []string {
	var keys []string
	for k, v := range m {
		if v == value {
			keys = append(keys, k)
		}
	}

	return keys
}

//CheckSaneDir Checks if the .sane directory exists. Creates it if it doesn't.
func CheckSaneDir() {
	home, err := homedir.Dir()
	util.Check(err)

	home = path.Join(home, "./.sane/")

	if _, err := os.Stat(home); os.IsNotExist(err) {
		// $HOME/.sane does not exist
		err := os.Mkdir(home, 777)
		util.Check(err)

		template := []byte("{\"repos\":[],\"aliases\":{}}")
		err = ioutil.WriteFile(path.Join(home, "./config.json"), template, 777)
		util.Check(err)
	}
}

//Read the config
func Read() SaneConfig {
	home, err := homedir.Dir()
	util.Check(err)

	repoFile := path.Join(home, "./.sane/config.json")

	var cfgStruct SaneConfig

	b, err := ioutil.ReadFile(repoFile)
	util.CheckWithMessage(err, "📭  Config file doesn't exist!")

	err = json.Unmarshal(b, &cfgStruct)
	util.CheckWithMessage(err, "😕  Invalid config file!")

	return cfgStruct
}

//Write the config
func Write(config SaneConfig) {
	home, err := homedir.Dir()
	util.Check(err)

	repoFile := path.Join(home, "./.sane/config.json")

	os.Remove(repoFile)
	f, _ := os.Create(repoFile)
	f.Close()

	b, err := json.Marshal(config)
	util.Check(err)

	err = ioutil.WriteFile(repoFile, b, os.ModePerm)
	util.Check(err)
}
