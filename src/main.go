package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"

	"github.com/mitchellh/go-homedir"
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
	Repos []Repo `json:"repos"`
}

var repoExp = regexp.MustCompile(`(?P<User>\w+)\/(?P<Name>\w+)(\/(?P<Branch>[\w\/]+))?(@(?P<Tag>[\w\.]+))?`)
var versionStr = "sane version 0.0.1"
var helpStr = `
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
var topicMap = map[string]string{
	"docker":    "🐳",
	"db":        "🗄",
	"server":    "📡",
	"browser":   "🌍",
	"neo4j":     "📊",
	"spring":    "🍃",
	"kafka":     "🐞",
	"couchbase": "🛋",
	"elk":       "📊 🔬 📺",
	"python":    "🐍",
	"c":         "𝗖",
	"cpp":       "𝗖++",
	"dotnet":    ".🌐",
	"java":      "☕️",
	"configs":   "📝",
	"json":      "J👶",
}

func getRepo(configString string) Repo {
	match := repoExp.FindStringSubmatch(configString)
	result := make(map[string]string)

	for i, name := range repoExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	var tag, branch string = "", ""

	if val, ok := result["Tag"]; ok && val != "" {
		tag = val
	}

	if val, ok := result["Branch"]; ok && val != "" {
		branch = val
	}

	return Repo{
		User:   result["User"],
		Name:   result["Name"],
		Branch: branch,
		Tag:    tag,
	}
}

func getConfig() SaneConfig {
	home, err := homedir.Dir()

	if err != nil {
		log.Fatal(err)
	}

	repoFile := path.Join(home, "./.sane/repos.json")
	repos, err := os.Open(repoFile)

	var cfgStruct SaneConfig
	b, err := ioutil.ReadAll(repos)
	err = json.Unmarshal(b, &cfgStruct)

	if err != nil {
		log.Fatal(err)
	}

	repos.Close()
	return cfgStruct
}

func writeConfig(config SaneConfig) {
	home, err := homedir.Dir()

	if err != nil {
		log.Fatal(err)
	}

	repoFile := path.Join(home, "./.sane/repos.json")
	repos, err := os.OpenFile(repoFile, os.O_WRONLY, os.ModePerm)

	b, err := json.Marshal(config)

	if err != nil {
		log.Fatal(err)
	}

	_, err = repos.Write(b)

	if err != nil {
		log.Fatal(err)
	}
}

func getTopics(topics []string) string {
	topicstr := ""
	for _, topic := range topics {
		if val, ok := topicMap[topic]; ok {
			topicstr += "[" + val + "] "
		}
	}

	return topicstr
}

//GhResult struct for result returned by Gh API
type GhResult struct {
	Names []string `json:"names"`
}

func getTopicsForRepo(user, name string) []string {
	client := &http.Client{}
	url := "https://api.github.com/repos/" + user + "/" + name + "/topics"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/vnd.github.mercy-preview+json")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	var ghResultStruct GhResult
	json.Unmarshal(b, &ghResultStruct)

	return ghResultStruct.Names
}

func pullRepo(repo Repo, home string, cfg SaneConfig) SaneConfig {
	var cmd = exec.Command("git", "clone", "https://github.com/"+repo.User+"/"+repo.Name+".git")
	var target = "./" + repo.User + "_" + repo.Name

	if repo.Tag != "" {
		target += "_" + repo.Tag
		cmd.Args = append(cmd.Args, "--branch")
		cmd.Args = append(cmd.Args, repo.Tag)
	} else {
		if repo.Branch != "" {
			target += "_" + repo.Branch
			cmd.Args = append(cmd.Args, "--branch")
			cmd.Args = append(cmd.Args, repo.Branch)
		}
	}

	cmd.Args = append(cmd.Args, path.Join(home, target))
	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	repo.Topics = getTopicsForRepo(repo.User, repo.Name)
	cfg.Repos = append(cfg.Repos, repo)
	writeConfig(cfg)

	return cfg
}

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	args := os.Args[1:]
	cfg := getConfig()

	home, err := homedir.Dir()

	if err != nil {
		log.Fatal(err)
	}

	home = path.Join(home, "./.sane/")

	if len(args) < 1 {
		fmt.Println("Expected at least one argument!")
		fmt.Println()
		fmt.Println(helpStr)
		return 1
	}

	command := args[0]

	if len(args) == 1 {
		switch command {
		case "-h", "--help":
			fmt.Println(helpStr)

		case "-v", "--version":
			fmt.Println(versionStr)

		case "list":
			for _, repo := range cfg.Repos {
				topics := getTopics(repo.Topics)

				branch := ""

				if repo.Branch == "" {
					if repo.Tag != "" {
						branch = "@" + repo.Tag
					}
				} else {
					branch = "@" + repo.Branch
				}

				if len(topics) != 0 {
					topics = "\n\t" + topics
				}

				fmt.Println("⚡️ " + repo.User + "/" + repo.Name + branch + topics)
			}

		case "list-topics":
			keys := make([]string, 0, len(topicMap))
			for k := range topicMap {
				keys = append(keys, k)
			}

			sort.Strings(keys)

			for _, k := range keys {
				fmt.Println(k + " => " + topicMap[k])
			}
		}

		return 0
	}

	repo := getRepo(args[1])

	switch command {
	case "get":
		//TODO: Check if config already contains this
		cfg = pullRepo(repo, home, cfg)
	case "start":
	case "apply":
	}

	return 0
}
