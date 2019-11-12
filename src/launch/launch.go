package launch

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"

	config "../config"
	repos "../repo"
	"gopkg.in/yaml.v2"
)

//StartContainerOrCompose start a container or a docker compose file
func StartContainerOrCompose(repo config.Repo, home string) {
	target := path.Join(home, repos.GetFolder(repo), "sane.yml")

	if _, err := os.Stat(target); os.IsNotExist(err) {
		fmt.Println("😐  Couldn't find sane.yml in " + target)
		os.Exit(1)
	}

	b, _ := ioutil.ReadFile(target)
	m := make(map[string]interface{})
	yaml.Unmarshal(b, &m)

	if val, ok := m["mode"]; ok {
		switch val.(string) {
		case "docker":
		//TODO
		case "docker-compose":
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
		default:
			fmt.Println("❌  Unsupported start mode \"" + val.(string) + "\"!")
			os.Exit(1)
		}
	} else {
		fmt.Println("❌  Config mode not set!")
		os.Exit(1)
	}
}
