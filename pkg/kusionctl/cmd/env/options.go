package env

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

type EnvOptions struct {
	envJson bool
}

func NewEnvOptions() *EnvOptions {
	return &EnvOptions{}
}

func (o *EnvOptions) Complete() {}

func (o *EnvOptions) Validate() error {
	return nil
}

func (o *EnvOptions) Run() error {
	env := []EnvVar{
		{Name: "KUSION_PATH", Value: os.Getenv("KUSION_PATH")},
	}

	if o.envJson {
		return printEnvAsJSON(env)
	}

	PrintEnv(os.Stdout, env)

	return nil
}

func printEnvAsJSON(env []EnvVar) error {
	m := make(map[string]string)
	for _, e := range env {
		if e.Name == "TERM" {
			continue
		}
		m[e.Name] = e.Value
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	if err := enc.Encode(m); err != nil {
		return err
	}

	return nil
}

// PrintEnv prints the environment variables to w.
func PrintEnv(w io.Writer, env []EnvVar) {
	for _, e := range env {
		if e.Name != "TERM" {
			switch runtime.GOOS {
			default:
				fmt.Fprintf(w, "%s=\"%s\"\n", e.Name, e.Value)
			case "plan9":
				if strings.IndexByte(e.Value, '\x00') < 0 {
					fmt.Fprintf(w, "%s='%s'\n", e.Name, strings.ReplaceAll(e.Value, "'", "''"))
				} else {
					v := strings.Split(e.Value, "\x00")
					fmt.Fprintf(w, "%s=(", e.Name)
					for x, s := range v {
						if x > 0 {
							fmt.Fprintf(w, " ")
						}
						fmt.Fprintf(w, "%s", s)
					}
					fmt.Fprintf(w, ")\n")
				}
			case "windows":
				fmt.Fprintf(w, "set %s=%s\n", e.Name, e.Value)
			}
		}
	}
}
