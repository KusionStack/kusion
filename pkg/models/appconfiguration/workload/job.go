package workload

// Job is a kind of workload profile that describes how to run your
// application code. This is typically used for tasks that take from a
// few senconds to a few days to complete.
type Job struct {
	*WorkloadBase `yaml:",inline" json:",inline"`

	// The scheduling strategy in Cron format.
	// More info: https://en.wikipedia.org/wiki/Cron.
	Schedule string `yaml:"schedule" json:"schedule"`
}
