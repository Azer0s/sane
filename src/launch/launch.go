package launch

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"

	config "../config"
	repos "../repo"
	"gopkg.in/yaml.v2"
)

func hasSaneYml(repo config.Repo, home string) string {
	target := path.Join(home, repos.GetFolder(repo), "sane.yml")

	if _, err := os.Stat(target); os.IsNotExist(err) {
		fmt.Println("😐  Couldn't find sane.yml in " + target)
		os.Exit(1)
	}

	return target
}

func startDockerCompose(m map[string]interface{}, repo config.Repo, home string) {
	if file, ok := m["file"]; ok {
		dockerComposeFile := path.Join(home, repos.GetFolder(repo), file.(string))

		cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "up")

		if scale, ok := m["scale"]; ok {
			cmd.Args = append(cmd.Args, "--scale")

			for _, v := range scale.([]interface{}) {
				for k, v1 := range v.(map[interface{}]interface{}) {
					cmd.Args = append(cmd.Args, k.(string)+"="+strconv.Itoa(v1.(int)))
				}
			}
		}

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		cmd.Run()
	} else {
		fmt.Println("❌  Docker compose file not set!")
		os.Exit(1)
	}
}

func startDocker(m map[string]interface{}, repo config.Repo, home string) {
	//TODO
}

//Start start a container or a docker compose file
func Start(repo config.Repo, home string) {
	target := hasSaneYml(repo, home)
	b, _ := ioutil.ReadFile(target)
	m := make(map[string]interface{})
	yaml.Unmarshal(b, &m)

	if val, ok := m["mode"]; ok {
		switch val.(string) {
		case "docker":
			startDocker(m, repo, home)
		case "docker-compose":
			startDockerCompose(m, repo, home)
		default:
			fmt.Println("❌  Unsupported start mode \"" + val.(string) + "\"!")
			os.Exit(1)
		}
	} else {
		fmt.Println("❌  Config mode not set!")
		os.Exit(1)
	}
}

func applyConfig(m map[string]interface{}, repo config.Repo, home string) {
	filesMap := make(map[string]string)

	if files, ok := m["files"]; ok {
		for _, v := range files.([]interface{}) {
			file := ""
			for k, v1 := range v.(map[interface{}]interface{}) {
				if k.(string) == "file" {
					file = k.(string)
				}

				if runtime.GOOS == k.(string) {
					filesMap[file] = v1.(string)
				}
			}
		}
	} else {
		fmt.Println("❌  Files not found!")
		os.Exit(1)
	}

	fmt.Println(filesMap)
}

func applyAliases(m map[string]interface{}, repo config.Repo, home string, cfg config.SaneConfig) {
	if aliases, ok := m["aliases"]; ok {
		for _, v := range aliases.([]interface{}) {
			for k, v1 := range v.(map[interface{}]interface{}) {
				cfg.Aliases[k.(string)] = v1.(string)
			}
		}

		fmt.Println("🎭  Writing aliases...")
		config.Write(cfg)
	} else {
		fmt.Println("❌  Aliases not found!")
		os.Exit(1)
	}
}

//Apply apply a config or a list of aliases
func Apply(repo config.Repo, home string, cfg config.SaneConfig) {
	target := hasSaneYml(repo, home)
	b, _ := ioutil.ReadFile(target)
	m := make(map[string]interface{})

	yaml.Unmarshal(b, &m)

	if val, ok := m["mode"]; ok {
		switch val.(string) {
		case "config":
			applyConfig(m, repo, home)
		case "aliases":
			applyAliases(m, repo, home, cfg)
		default:
			fmt.Println("❌  Unsupported start mode \"" + val.(string) + "\"!")
			os.Exit(1)
		}
	} else {
		fmt.Println("❌  Config mode not set!")
		os.Exit(1)
	}
}
