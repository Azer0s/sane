package launch

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"

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

func couldntParse(info string) {
	fmt.Println("❌  Couldn't parse config! " + info)
	os.Exit(1)
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
	configs := extractDockerConfig(m)

	sort.SliceStable(configs, func(i, j int) bool {
		return configs[i].Start < configs[j].Start
	})

	for _, dockerConfig := range configs {
		command := "docker run"

		if dockerConfig.Deamon {
			command += " -d"
		}

		command += " --name " + dockerConfig.Name

		if dockerConfig.Net != "" {
			command += " --net " + dockerConfig.Net
		}

		if dockerConfig.Ipc != "" {
			command += " --ipc " + dockerConfig.Ipc
		}

		if dockerConfig.Pid != "" {
			command += " --pid " + dockerConfig.Pid
		}

		for _, port := range dockerConfig.Ports {
			command += " -p " + strconv.Itoa(port.Source) + ":" + strconv.Itoa(port.Target)
		}

		for _, env := range dockerConfig.Environment {
			command += " --env \"" + env.Key + "=" + env.Value + "\""
		}

		command += " " + dockerConfig.Image

		//TODO: Run command, add stop command
		fmt.Println(command)
	}
}

func extractDockerConfig(m map[string]interface{}) []DockerConfig {
	dockerConfigs := make([]DockerConfig, 0)
	configs := m["containers"].(map[interface{}]interface{})

	for k, v := range configs {
		vals := v.(map[interface{}]interface{})
		cfg := DockerConfig{
			Name:        k.(string),
			Ports:       make([]PortMapping, 0),
			Environment: make([]EnvironmentPair, 0),
			Start:       math.MaxInt32,
		}

		if deamon, ok := vals["deamon"]; ok {
			cfg.Deamon = deamon.(bool)
		}

		if net, ok := vals["net"]; ok {
			cfg.Net = net.(string)
		}

		if ipc, ok := vals["ipc"]; ok {
			cfg.Ipc = ipc.(string)
		}

		if pid, ok := vals["pid"]; ok {
			cfg.Pid = pid.(string)
		}

		if image, ok := vals["image"]; ok {
			cfg.Image = image.(string)
		} else {
			couldntParse("Image not specified (" + cfg.Name + ")!")
		}

		if start, ok := vals["start"]; ok {
			cfg.Start = start.(int)
		}

		if env, ok := vals["environment"]; ok {
			for _, e := range env.([]interface{}) {
				for envK, envV := range e.(map[interface{}]interface{}) {
					cfg.Environment = append(cfg.Environment, EnvironmentPair{
						Key:   fmt.Sprintf("%v", envK),
						Value: fmt.Sprintf("%v", envV),
					})
				}
			}
		}

		if port, ok := vals["ports"]; ok {
			for _, v := range port.([]interface{}) {
				ports := strings.Split(v.(string), ":")
				source, err := strconv.Atoi(ports[0])
				if err != nil {
					couldntParse("")
				}

				target, err := strconv.Atoi(ports[1])

				if err != nil {
					couldntParse("")
				}

				cfg.Ports = append(cfg.Ports, PortMapping{
					Source: source,
					Target: target,
				})
			}
		}

		dockerConfigs = append(dockerConfigs, cfg)
	}

	return dockerConfigs
}

//Start start a container or a docker compose file
func Start(repo config.Repo, home string) {
	target := hasSaneYml(repo, home)
	b, _ := ioutil.ReadFile(target)
	m := make(map[string]interface{})
	err := yaml.Unmarshal(b, &m)

	if err != nil {
		couldntParse("")
	}

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

	err := yaml.Unmarshal(b, &m)

	if err != nil {
		couldntParse("")
	}

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
