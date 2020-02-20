package src

//DockerConfig the docker run configuration
type DockerConfig struct {
	Name        string
	Deamon      bool
	Net         string
	Ipc         string
	Pid         string
	Ports       []PortMapping
	Volumes     []VolumeMapping
	Environment []EnvironmentPair
	Image       string
	Start       int
	Stop        int
}

//EnvironmentPair a k-v pair for an environment variable
type EnvironmentPair struct {
	Key   string
	Value string
}

//PortMapping a k-v pair for an docker port mapping
type PortMapping struct {
	Source string
	Target string
}

//VolumeMapping a k-v pair for an docker volume mapping
type VolumeMapping struct {
	Source string
	Target string
}
