package appconfiguration

import (
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"kusionstack.io/kusion/pkg/models"
)

// CallGeneratorFuncs calls each NewGeneratorFunc in the given slice
// and returns a slice of AppsGenerator instances.
func CallGeneratorFuncs(newGenerators ...NewGeneratorFunc) ([]Generator, error) {
	gs := make([]Generator, 0, len(newGenerators))
	for _, newGenerator := range newGenerators {
		if g, err := newGenerator(); err != nil {
			return nil, err
		} else {
			gs = append(gs, g)
		}
	}
	return gs, nil
}

// CallGenerators calls the Generate method of each AppsGenerator
// instance returned by the given NewGeneratorFuncs.
func CallGenerators(spec *models.Spec, newGenerators ...NewGeneratorFunc) error {
	gs, err := CallGeneratorFuncs(newGenerators...)
	if err != nil {
		return err
	}
	for _, g := range gs {
		if err := g.Generate(spec); err != nil {
			return err
		}
	}
	return nil
}

// ForeachOrdered executes the given function on each
// item in the map in order of their keys.
func ForeachOrdered[T any](m map[string]T, f func(key string, value T) error) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		if err := f(k, v); err != nil {
			return err
		}
	}

	return nil
}

// GenericPtr returns a pointer to the provided value.
func GenericPtr[T any](i T) *T {
	return &i
}

// MergeMaps merges multiple map[string]string into one
// map[string]string.
// If a map is nil, it skips it and moves on to the next one. For each
// non-nil map, it iterates over its key-value pairs and adds them to
// the merged map. Finally, it returns the merged map.
func MergeMaps(maps ...map[string]string) map[string]string {
	merged := make(map[string]string)

	for _, m := range maps {
		if m == nil {
			continue
		}
		for k, v := range m {
			merged[k] = v
		}
	}

	return merged
}

// KubernetesResourceID returns the unique ID of a Kubernetes resource
// based on its type and metadata.
func KubernetesResourceID(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:code-city:code-citydev
	id := typeMeta.APIVersion + ":" + typeMeta.Kind + ":"
	if objectMeta.Namespace != "" {
		id += objectMeta.Namespace + ":"
	}
	id += objectMeta.Name
	return id
}

// AppendToSpec adds a Kubernetes resource to a spec's resources
// slice.
func AppendToSpec(resourceType models.Type, resourceID string, spec *models.Spec, resource any) error {
	unstructured, err := runtime.DefaultUnstructuredConverter.ToUnstructured(resource)
	if err != nil {
		return err
	}
	r := models.Resource{
		ID:         resourceID,
		Type:       resourceType,
		Attributes: unstructured,
		DependsOn:  nil,
		Extensions: nil,
	}
	spec.Resources = append(spec.Resources, r)
	return nil
}

// UniqueAppName returns a unique name for a workload based on its
// project and app name.
func UniqueAppName(projectName, stackName, appName string) string {
	return projectName + "-" + stackName + "-" + appName
}

// UniqueAppLabels returns a map of labels that identify an app based
// on its project and name.
func UniqueAppLabels(projectName, appName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/part-of": projectName,
		"app.kubernetes.io/name":    appName,
	}
}

// MagicEnvVar generates a specialized EnvVar based on the key and
// value of environment.
func MagicEnvVar(k, v string) *corev1.EnvVar {
	supportedParsers := []MagicEnvParser{
		SecretEnvParser,
		ConfigMapEnvParser,
		RawEnvParser,
	}
	for _, p := range supportedParsers {
		if p.Match(k, v) {
			return p.Gen(k, v)
		}
	}
	return nil
}

var (
	SecretEnvParser    MagicEnvParser = NewSecretEnvParser()
	ConfigMapEnvParser                = NewConfigMapEnvParser()
	RawEnvParser                      = NewRawEnvParser()
)

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
