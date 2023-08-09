package container

type Container struct {
	// Image to run for this container
	Image string `yaml:"image,omitempty" json:"image,omitempty"`
	// Entrypoint array.
	// The image's ENTRYPOINT is used if this is not provided.
	Command []string `yaml:"command,omitempty" json:"command,omitempty"`
	// Arguments to the entrypoint.
	// The image's CMD is used if this is not provided.
	Args []string `yaml:"args,omitempty" json:"args,omitempty"`
	// Collection of environment variables to set in the container.
	// The value of environment variable may be static text or a value from a secret.
	Env map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	// The current working directory of the running process defined in entrypoint.
	WorkingDir string `yaml:"workingDir,omitempty" json:"workingDir,omitempty"`
}
