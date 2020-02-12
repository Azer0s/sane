package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"

	config "./config"
	launch "./launch"
	repos "./repo"

	"github.com/mitchellh/go-homedir"
)

var versionStr = "sane version 1.0.0"
var helpStr = `
sane - A package manager for sane configuration
https://github.com/Azer0s/sane

Flags:
  -v --version		Displays the program version string.
  -h --help   		Displays the help page.
  
Commands:
  get <config>  	Pull a config from GitHub.
  purge <config>	Purge a pulled config from disk.

  start <config>	Starts an application specified by a sanefile.
  stop <config>		Stops an application specified by a sanefile.

  apply <config>	Applies a configuration specified by a sanefile.
  remove <config>	Removes a configuration specified by a sanefile.

  list        		Lists available configs.
  aliases       	Lists all aliases.
  alias <config> <name>	Alias a config.
  rmaliases        	Remove all aliases.
  dealias <config>	Remove alias from a config.
`

func main() {
	err := exec.Command("docker", "-v").Run()
	if err != nil {
		fmt.Println("🐳❌  Docker not installed!")
		os.Exit(1)
	}

	err = exec.Command("docker", "info").Run()
	if err != nil {
		fmt.Println("👻❌  Docker not reachable. Is the docker deamon running?")
		os.Exit(1)
	}

	err = exec.Command("docker-compose", "version").Run()
	if err != nil {
		fmt.Println("👷‍❌  Docker-compose not installed!")
		os.Exit(1)
	}

	args := os.Args[1:]
	cfg := config.Read()

	home, err := homedir.Dir()

	if err != nil {
		log.Fatal(err)
	}

	home = path.Join(home, "./.sane/")

	if len(args) < 1 {
		fmt.Println("Expected at least one argument!")
		fmt.Println()
		fmt.Println(helpStr)
		os.Exit(1)
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
				topics := repos.GetTopics(repo.Topics)

				branch := ""

				if repo.Branch == "" {
					if repo.Tag != "" {
						branch = "@" + repo.Tag
					}
				} else {
					branch = "/" + repo.Branch
				}

				if len(topics) != 0 {
					topics = "\n\t" + topics
				}

				fmt.Println("⚡️ " + repo.User + "/" + repo.Name + branch + topics)
			}

		case "aliases":
			for k, v := range cfg.Aliases {
				fmt.Println("🎭  " + k + " => " + v)
			}

		case "rmaliases":
			cfg.Aliases = make(map[string]string)
			config.Write(cfg)

		default:
			fmt.Println("🤷 ❌ Command unrecognized!‍")
			fmt.Println(helpStr)
			os.Exit(1)
		}

		os.Exit(0)
	}

	var repo config.Repo

	if regexp.MustCompile(`^[\w-]+$`).MatchString(args[1]) {
		if val, ok := cfg.Aliases[args[1]]; ok {
			repo = repos.GetRepoFromString(val)
		} else {
			fmt.Println("🤫 ❌  Alias " + args[1] + " not found!")
			os.Exit(1)
		}
	} else {
		repo = repos.GetRepoFromString(args[1])
	}

	switch command {
	case "get":
		cfg = repos.Pull(repo, home, cfg)
		fmt.Println("😊  New config ready to use!")
	case "purge":
		cfg = repos.Purge(repo, home, cfg)
		fmt.Println("😬  Config successfully removed!")
		config.Write(cfg)
	case "start":
		cfg = repos.AutoPull(cfg, repo, home)
		fmt.Println("🚀  Starting " + args[1] + "...")
		launch.Start(repo, home)
	case "stop":
		fmt.Println("✋  Stopping " + args[1] + "...")
		launch.Stop(repo, home)
	case "apply":
		cfg = repos.AutoPull(cfg, repo, home)
		fmt.Println("✍️  ​Applying config " + args[1] + "...")
		launch.Do(repo, home, cfg, config.APPLY)
	case "remove":
		cfg = repos.AutoPull(cfg, repo, home)
		fmt.Println("💣  Removing config... ")
		launch.Do(repo, home, cfg, config.REMOVE)
	case "alias":
		fmt.Println("🤫  Aliasing " + args[1] + " to " + args[2])
		cfg.Aliases[args[2]] = args[1]
		config.Write(cfg)
	case "dealias":
		fmt.Println("👀  Removing alias to " + args[1])

		keys := config.Mapkeys(cfg.Aliases, args[1])

		for _, key := range keys {
			delete(cfg.Aliases, key)
		}

		config.Write(cfg)
	}

	os.Exit(0)
}
