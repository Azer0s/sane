package src

import (
	"encoding/json"
	"io/ioutil"
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

//CheckSaneDir Checks if the .sane directory exists. Creates it if it doesn't.
func CheckSaneDir() {
	home, err := homedir.Dir()
	Check(err)

	home = path.Join(home, "./.sane/")

	if _, err := os.Stat(home); os.IsNotExist(err) {
		// $HOME/.sane does not exist
		err := os.Mkdir(home, 0777)
		Check(err)

		template := []byte("{\"repos\":[],\"aliases\":{}}")
		err = ioutil.WriteFile(path.Join(home, "./config.json"), template, 0777)
		Check(err)
	}
}

//ReadConfig the config
func ReadConfig() SaneConfig {
	home, err := homedir.Dir()
	Check(err)

	repoFile := path.Join(home, "./.sane/config.json")

	var cfgStruct SaneConfig

	b, err := ioutil.ReadFile(repoFile)
	CheckWithMessage(err, "📭  Config file doesn't exist!")

	err = json.Unmarshal(b, &cfgStruct)
	CheckWithMessage(err, "😕  Invalid config file!")

	return cfgStruct
}

//WriteConfig the config
func WriteConfig(config SaneConfig) {
	home, err := homedir.Dir()
	Check(err)

	repoFile := path.Join(home, "./.sane/config.json")

	_ = os.Remove(repoFile)
	f, _ := os.Create(repoFile)
	_ = f.Close()

	b, err := json.Marshal(config)
	Check(err)

	err = ioutil.WriteFile(repoFile, b, os.ModePerm)
	Check(err)
}
