package launch

import (
	"fmt"
	"github.com/docker/docker/pkg/fileutils"
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

func startDocker(m map[string]interface{}) {
	configs := extractDockerConfig(m)

	started := make([]DockerConfig, 0)

	sort.SliceStable(configs, func(i, j int) bool {
		return configs[i].Start < configs[j].Start
	})

	for _, dockerConfig := range configs {
		cmd := exec.Command("docker", "run")

		if dockerConfig.Deamon {
			cmd.Args = append(cmd.Args, "-d")
		}

		cmd.Args = append(cmd.Args, "--name")
		cmd.Args = append(cmd.Args, dockerConfig.Name)

		if dockerConfig.Net != "" {
			cmd.Args = append(cmd.Args, "--net")
			cmd.Args = append(cmd.Args, dockerConfig.Net)
		}

		if dockerConfig.Ipc != "" {
			cmd.Args = append(cmd.Args, "--ipc")
			cmd.Args = append(cmd.Args, dockerConfig.Ipc)
		}

		if dockerConfig.Pid != "" {
			cmd.Args = append(cmd.Args, "--pid")
			cmd.Args = append(cmd.Args, dockerConfig.Pid)
		}

		for _, port := range dockerConfig.Ports {
			cmd.Args = append(cmd.Args, "-p")
			cmd.Args = append(cmd.Args, strconv.Itoa(port.Source)+":"+strconv.Itoa(port.Target))
		}

		for _, env := range dockerConfig.Environment {
			cmd.Args = append(cmd.Args, "--env")

			if strings.Contains(env.Value, " ") {
				env.Value = "\"" + env.Value + "\""
			}

			cmd.Args = append(cmd.Args, env.Key+"="+env.Value)
		}

		cmd.Args = append(cmd.Args, dockerConfig.Image)

		fmt.Println("🐳  Starting container '" + dockerConfig.Name + "'...")
		err := cmd.Run()

		if err != nil {
			fmt.Println("❌  There was an error while starting the container! Rolling back...")
			for _, s := range started {
				exec.Command("docker", "stop", s.Name).Run()
				exec.Command("docker", "rm", s.Name).Run()
			}
			os.Exit(1)
		}

		started = append(started, dockerConfig)
	}
}

func stopDocker(m map[string]interface{}) {
	configs := extractDockerConfig(m)

	sort.SliceStable(configs, func(i, j int) bool {
		return configs[i].Stop < configs[j].Stop
	})

	for _, dockerConfig := range configs {
		fmt.Println("🐳  Stopping container '" + dockerConfig.Name + "'...")
		err1 := exec.Command("docker", "stop", dockerConfig.Name).Run()
		err2 := exec.Command("docker", "rm", dockerConfig.Name).Run()

		if err1 != nil || err2 != nil {
			fmt.Println("❌  There was an error while stopping the container!")
			os.Exit(1)
		}
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
			Stop:        math.MaxInt32,
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

		if stop, ok := vals["stop"]; ok {
			cfg.Stop = stop.(int)
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
			startDocker(m)
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

//Stop stop a container
func Stop(repo config.Repo, home string) {
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
			stopDocker(m)
		default:
			fmt.Println("❌  Unsupported stop mode \"" + val.(string) + "\"!")
			os.Exit(1)
		}
	} else {
		fmt.Println("❌  Config mode not set!")
		os.Exit(1)
	}
}

func extractFileConfig(m map[string]interface{}) map[string]string {
	fileMap := make(map[string]string)

	if files, ok := m["files"]; ok {
		for _, v := range files.([]interface{}) {
			m := v.(map[interface{}]interface{})
			fileMap[m["file"].(string)] = os.ExpandEnv(m[runtime.GOOS].(string))
		}
	} else {
		fmt.Println("❌  Files not found!")
		os.Exit(1)
	}

	return fileMap
}

func applyConfig(m map[string]interface{}, repo config.Repo, home string) {
	files := extractFileConfig(m)
	for src, dst := range files {
		err := os.Rename(dst, dst+".backup")

		if err != nil {
			fmt.Println("❌  There was an error while moving a file!")
			os.Exit(1)
		}

		target := path.Join(home, repos.GetFolder(repo), src)

		_, err = fileutils.CopyFile(target, dst)
		if err != nil {
			fmt.Println("❌  There was an error while moving a file!")
			os.Exit(1)
		}
	}
}

func removeConfig(m map[string]interface{}) {
	files := extractFileConfig(m)
	for _, dst := range files {
		err := os.Remove(dst)

		if err != nil {
			fmt.Println("❌  There was an error while deleting a file!")
			os.Exit(1)
		}
		err = os.Rename(dst+".backup", dst)
		if err != nil {
			fmt.Println("❌  There was an error while moving a file!")
			os.Exit(1)
		}
	}
}

func applyAliases(m map[string]interface{}, cfg config.SaneConfig) {
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

func removeAliases(m map[string]interface{}, cfg config.SaneConfig) {
	if aliases, ok := m["aliases"]; ok {
		for _, v := range aliases.([]interface{}) {
			for k, _ := range v.(map[interface{}]interface{}) {
				delete(cfg.Aliases, k.(string))
			}
		}

		fmt.Println("🎭  Writing aliases...")
		config.Write(cfg)
	} else {
		fmt.Println("❌  Aliases not found!")
		os.Exit(1)
	}
}

//Do apply/remove a config or a list of aliases
func Do(repo config.Repo, home string, cfg config.SaneConfig, mode string) {
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
			if mode == "apply" {
				applyConfig(m, repo, home)
			} else {
				removeConfig(m)
			}
		case "aliases":
			if mode == "apply" {
				applyAliases(m, cfg)
			} else {
				removeAliases(m, cfg)
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
