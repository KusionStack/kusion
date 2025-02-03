package options

type ApplyOptions interface {
	PreviewOptions

	GetYes() bool
	GetDryRun() bool
	GetWatch() bool
	GetTimeout() int
	GetPortForward() int
}
