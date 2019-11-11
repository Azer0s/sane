package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"

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

var repoExp = regexp.MustCompile(`(?P<User>\w+)\/(?P<Name>\w+)\/(?P<Branch>[\w\/]+)(@(?P<Tag>[\w\.]+))?`)

func getRepo(configString string) Repo {
	match := repoExp.FindStringSubmatch(configString)
	result := make(map[string]string)

	for i, name := range repoExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	var tag string = ""

	if val, ok := result["Tag"]; ok {
		tag = val
	}

	return Repo{
		User:   result["User"],
		Name:   result["Name"],
		Branch: result["Branch"],
		Tag:    tag,
	}
}

func mainReturnWithCode() int {
	args := os.Args[1:]

	home, err := homedir.Dir()

	if err != nil {
		log.Fatal(err)
	}

	repoFile := path.Join(home, "./.sane/repos.json")
	repos, err := os.Open(repoFile)

	var cfgStruct SaneConfig

	if err != nil {
		os.Mkdir(path.Join(home, "./.sane/"), os.ModePerm)

		f, err := os.Create(repoFile)

		if err != nil {
			log.Fatal(err)
		}

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
		json.Unmarshal(b, &cfgStruct)
	}

	fmt.Println(cfgStruct)

	helpStr := `
sane - A package manager for sane configuration
https://github.com/Azer0s/sane

Flags:
  -v --version		Displays the program version string.
  -h --help   		Displays the help page.
  
Commands:
  start <config>	Starts a docker container with the specified config.
  stop <config>		Stops a docker container with the specified config.
  list        		Lists available configs.
  apply <config>	Applies a configuration to the home directory.
  remove <config>	Removes a configuration from the home directory.
`

	if len(args) < 2 {
		fmt.Println(helpStr)
		return 1
	}

	command := args[0]

	switch command {
	case "start":
		fmt.Println("Start")

	case "list":

	case "apply":

	}

	return 0
}
