package launch

type DockerConfig struct {
	Name        string
	Deamon      bool
	Net         string
	Ipc         string
	Pid         string
	Ports       []PortMapping
	Environment []EnvironmentPair
	Image       string
	Start       int
	Stop        int
}

type EnvironmentPair struct {
	Key   string
	Value string
}

type PortMapping struct {
	Source int
	Target int
}
