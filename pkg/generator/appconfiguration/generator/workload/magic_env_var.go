package workload

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

var (
	SecretEnvParser    MagicEnvParser = NewSecretEnvParser()
	ConfigMapEnvParser                = NewConfigMapEnvParser()
	RawEnvParser                      = NewRawEnvParser()

	supportedParsers = []MagicEnvParser{
		SecretEnvParser,
		ConfigMapEnvParser,
		// As the default parser, the RawEnvParser should be placed at
		// the end.
		RawEnvParser,
	}
)

// MagicEnvVar generates a specialized EnvVar based on the key and
// value of environment.
//
// Examples:
//
//	MagicEnvVar("secret_key", "secret://my_secret/my_key")
//	MagicEnvVar("config_key", "configmap://my_configmap/my_key")
//	MagicEnvVar("key", "value")
func MagicEnvVar(k, v string) *corev1.EnvVar {
	for _, p := range supportedParsers {
		if p.Match(k, v) {
			return p.Gen(k, v)
		}
	}
	return nil
}

// MagicEnvParser is an interface for environment variable parsers.
type MagicEnvParser interface {
	Match(k, v string) (matched bool)
	Gen(k, v string) *corev1.EnvVar
}

// rawEnvParser is a parser for raw environment variables.
type rawEnvParser struct{}

// NewRawEnvParser creates a new instance of RawEnvParser.
func NewRawEnvParser() MagicEnvParser {
	return &rawEnvParser{}
}

// Match checks if the value matches the raw parser.
func (*rawEnvParser) Match(_ string, _ string) bool {
	return true
}

// Gen generates a raw environment variable.
func (*rawEnvParser) Gen(k string, v string) *corev1.EnvVar {
	return &corev1.EnvVar{
		Name:  k,
		Value: v,
	}
}

// secretEnvParser is a parser for secret-based environment variables.
type secretEnvParser struct {
	prefix string
}

// NewSecretEnvParser creates a new instance of SecretEnvParser.
func NewSecretEnvParser() MagicEnvParser {
	return &secretEnvParser{
		prefix: "secret://",
	}
}

// Match checks if the value matches the secret parser.
func (p *secretEnvParser) Match(_ string, v string) bool {
	return strings.HasPrefix(v, p.prefix)
}

// Gen generates a secret-based environment variable.
func (p *secretEnvParser) Gen(k string, v string) *corev1.EnvVar {
	vv := strings.TrimPrefix(v, p.prefix)
	vs := strings.Split(vv, "/")
	if len(vs) != 2 {
		return nil
	}

	return &corev1.EnvVar{
		Name: k,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: vs[0],
				},
				Key: vs[1],
			},
		},
	}
}

// configMapEnvParser is a parser for configmap-based environment
// variables.
type configMapEnvParser struct {
	prefix string
}

// NewConfigMapEnvParser creates a new instance of ConfigMapEnvParser.
func NewConfigMapEnvParser() MagicEnvParser {
	return &configMapEnvParser{
		prefix: "configmap://",
	}
}

// Match checks if the value matches the configmap parser.
func (p *configMapEnvParser) Match(_ string, v string) bool {
	return strings.HasPrefix(v, p.prefix)
}

// Gen generates a configmap-based environment variable.
func (p *configMapEnvParser) Gen(k string, v string) *corev1.EnvVar {
	vv := strings.TrimPrefix(v, p.prefix)
	vs := strings.Split(vv, "/")
	if len(vs) != 2 {
		return nil
	}

	return &corev1.EnvVar{
		Name: k,
		ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: vs[0],
				},
				Key: vs[1],
			},
		},
	}
}
