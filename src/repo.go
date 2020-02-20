package src

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
)

//TopicMap a list of topics and corresponding emojis
var TopicMap = map[string]string{
	"docker":    "🐳",
	"db":        "🗄",
	"server":    "🛰",
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

var repoExp = regexp.MustCompile(`^(?P<User>\w+)/(?P<Name>\w+)((/(?P<Branch>[\w/\-_]+))?|(@(?P<Tag>[\w.]+))?)$`)

//GetRepoFromString get repo config from string
func GetRepoFromString(configString string) Repo {
	match := repoExp.FindStringSubmatch(configString)

	if match == nil {
		fmt.Println("❌  Invalid repo format!")
		os.Exit(1)
	}

	result := make(map[string]string)

	for i, name := range repoExp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	var tag, branch = "", ""

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

//GetTopicEmojis get emoji representation of Repo topics
func GetTopicEmojis(topics []string) string {
	topicstr := ""
	for _, topic := range topics {
		if val, ok := TopicMap[topic]; ok {
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
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/vnd.github.mercy-preview+json")

	resp, err := client.Do(req)
	Check(err)

	b, _ := ioutil.ReadAll(resp.Body)
	var ghResultStruct GhResult
	_ = json.Unmarshal(b, &ghResultStruct)

	return ghResultStruct.Names
}

//PullRepo pull a repo from Gh
func PullRepo(repo Repo, home string, cfg SaneConfig) SaneConfig {
	var cmd = exec.Command("git", "clone", "https://github.com/"+repo.User+"/"+repo.Name+".git")

	if repo.Tag != "" {
		cmd.Args = append(cmd.Args, "--branch")
		cmd.Args = append(cmd.Args, repo.Tag)
	} else {
		if repo.Branch != "" {
			cmd.Args = append(cmd.Args, "--branch")
			cmd.Args = append(cmd.Args, repo.Branch)
		}
	}

	target := path.Join(home, GetRepoFolder(repo))

	if _, err := os.Stat(target); !os.IsNotExist(err) {
		cfg = PurgeRepo(repo, home, cfg)
	}

	cmd.Args = append(cmd.Args, target)

	fmt.Println("🌍  Downloading repo...")
	err := cmd.Run()
	Check(err)

	repo.Topics = getTopicsForRepo(repo.User, repo.Name)
	cfg.Repos = append(cfg.Repos, repo)

	fmt.Println("📝  Registering new config...")
	WriteConfig(cfg)

	return cfg
}

//GetRepoFolder get the folder of a config
func GetRepoFolder(repo Repo) string {
	target := "./" + repo.User + "_" + repo.Name

	if repo.Tag != "" {
		target += "_" + repo.Tag
	} else {
		if repo.Branch != "" {
			target += "_" + repo.Branch
		}
	}

	return target
}

//PurgeRepo purge a repo
func PurgeRepo(repo Repo, home string, cfg SaneConfig) SaneConfig {
	target := path.Join(home, GetRepoFolder(repo))

	fmt.Println("🗑  ​Purging repo...")
	_ = exec.Command("rm", "-rf", target).Run()

	if Contains(cfg.Repos, repo) {
		i := IndexOf(cfg.Repos, repo)
		cfg.Repos = append(cfg.Repos[:i], cfg.Repos[i+1:]...)
	}

	return cfg
}

//AutoPullRepo pulls if not exists
func AutoPullRepo(cfg SaneConfig, repo Repo, home string) SaneConfig {
	if !Contains(cfg.Repos, repo) {
		fmt.Println("🤷‍  Config missing, pulling automatically...")
		cfg = PullRepo(repo, home, cfg)
	}

	return cfg
}
