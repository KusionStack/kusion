package workload

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"golang.org/x/exp/maps"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/workload/secret"
	"kusionstack.io/kusion/pkg/util/net"
	"kusionstack.io/kusion/pkg/workspace"
)

type Generator struct {
	// Project represents the Project name
	Project string
	// Stack represents the Stack name
	Stack string
	// App represents the application name
	App string
	// Namespace represents the K8s Namespace
	Namespace string
	// Workload represents the Workload configuration
	Workload *v1.Workload
	// PlatformConfigs represents the module platform configurations
	PlatformConfigs map[string]v1.GenericConfig
	// SecretStoreSpec contains configuration to describe target secret store.
	SecretStoreSpec *v1.SecretStore
}

func NewWorkloadGeneratorFunc(g *Generator) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		if len(g.Project) == 0 {
			return nil, fmt.Errorf("project name must not be empty")
		}

		if len(g.Stack) == 0 {
			return nil, fmt.Errorf("stack name must not be empty")
		}

		if len(g.App) == 0 {
			return nil, fmt.Errorf("app name must not be empty")
		}

		return g, nil
	}
}

func (g *Generator) Generate(spec *v1.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(v1.Resources, 0)
	}

	if g.Workload != nil {
		var gfs []modules.NewGeneratorFunc

		switch g.Workload.Header.Type {
		case v1.TypeService:
			gfs = append(gfs, NewWorkloadServiceGeneratorFunc(g), secret.NewSecretGeneratorFunc(&secret.GeneratorRequest{
				Project:     g.Project,
				Namespace:   g.Namespace,
				Workload:    g.Workload,
				SecretStore: g.SecretStoreSpec,
			}))
		case v1.TypeJob:
			gfs = append(gfs, NewJobGeneratorFunc(g), secret.NewSecretGeneratorFunc(&secret.GeneratorRequest{
				Project:     g.Project,
				Namespace:   g.Namespace,
				Workload:    g.Workload,
				SecretStore: g.SecretStoreSpec,
			}))
		}

		if err := modules.CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}

	return nil
}

func toOrderedContainers(
	appContainers map[string]v1.Container,
	uniqueAppName string,
) ([]corev1.Container, []corev1.Volume, []corev1.ConfigMap, error) {
	// Create a slice of containers based on the App's containers.
	var containers []corev1.Container

	// Create a slice of volumes and configMaps based on the containers' files to be created.
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var configMaps []corev1.ConfigMap

	if err := modules.ForeachOrdered(appContainers, func(containerName string, c v1.Container) error {
		// Create a slice of env vars based on the container's env vars.
		var envs []corev1.EnvVar
		for _, m := range c.Env {
			envs = append(envs, *MagicEnvVar(m.Key.(string), m.Value.(string)))
		}

		resourceRequirements, err := handleResourceRequirementsV1(c.Resources)
		if err != nil {
			return err
		}

		// Create a container object.
		ctn := corev1.Container{
			Name:       containerName,
			Image:      c.Image,
			Command:    c.Command,
			Args:       c.Args,
			WorkingDir: c.WorkingDir,
			Env:        envs,
			Resources:  resourceRequirements,
		}
		if err = updateContainer(&c, &ctn); err != nil {
			return err
		}

		// Append the configMap, volume and volumeMount objects into the corresponding slices.
		volumes, volumeMounts, configMaps, err = handleFileCreation(c, uniqueAppName, containerName)
		if err != nil {
			return err
		}
		ctn.VolumeMounts = append(ctn.VolumeMounts, volumeMounts...)

		// Append more volumes and volumeMounts
		otherVolumes, otherVolumeMounts, err := handleDirCreation(c)
		if err != nil {
			return err
		}
		volumes = append(volumes, otherVolumes...)
		ctn.VolumeMounts = append(ctn.VolumeMounts, otherVolumeMounts...)

		// Append the container object to the containers slice.
		containers = append(containers, ctn)
		return nil
	}); err != nil {
		return nil, nil, nil, err
	}
	return containers, volumes, configMaps, nil
}

// updateContainer updates corev1.Container with passed parameters.
func updateContainer(in *v1.Container, out *corev1.Container) error {
	if in.ReadinessProbe != nil {
		readinessProbe, err := convertKusionProbeToV1Probe(in.ReadinessProbe)
		if err != nil {
			return err
		}
		out.ReadinessProbe = readinessProbe
	}

	if in.LivenessProbe != nil {
		livenessProbe, err := convertKusionProbeToV1Probe(in.LivenessProbe)
		if err != nil {
			return err
		}
		out.LivenessProbe = livenessProbe
	}

	if in.StartupProbe != nil {
		startupProbe, err := convertKusionProbeToV1Probe(in.StartupProbe)
		if err != nil {
			return err
		}
		out.StartupProbe = startupProbe
	}

	if in.Lifecycle != nil {
		lifecycle, err := convertKusionLifecycleToV1Lifecycle(in.Lifecycle)
		if err != nil {
			return err
		}
		out.Lifecycle = lifecycle
	}

	return nil
}

// handleResourceRequirementsV1 parses the resources parameter if specified and
// returns ResourceRequirements.
func handleResourceRequirementsV1(resources map[string]string) (corev1.ResourceRequirements, error) {
	result := corev1.ResourceRequirements{}
	if resources == nil {
		return result, nil
	}
	for key, value := range resources {
		resourceName := corev1.ResourceName(key)
		requests, limits, err := populateResourceLists(resourceName, value)
		if err != nil {
			return result, err
		}
		if requests != nil && result.Requests == nil {
			result.Requests = make(corev1.ResourceList)
		}
		maps.Copy(result.Requests, requests)
		if limits != nil && result.Limits == nil {
			result.Limits = make(corev1.ResourceList)
		}
		maps.Copy(result.Limits, limits)
	}
	return result, nil
}

// populateResourceLists takes strings of form <resourceName>=[<minValue>-]<maxValue> and
// returns request&limit ResourceList.
func populateResourceLists(name corev1.ResourceName, spec string) (corev1.ResourceList, corev1.ResourceList, error) {
	requests := corev1.ResourceList{}
	limits := corev1.ResourceList{}

	parts := strings.Split(spec, "-")
	if len(parts) == 1 {
		resourceQuantity, err := resource.ParseQuantity(parts[0])
		if err != nil {
			return nil, nil, err
		}
		limits[name] = resourceQuantity
	} else if len(parts) == 2 {
		resourceQuantity, err := resource.ParseQuantity(parts[0])
		if err != nil {
			return nil, nil, err
		}
		requests[name] = resourceQuantity
		resourceQuantity, err = resource.ParseQuantity(parts[1])
		if err != nil {
			return nil, nil, err
		}
		limits[name] = resourceQuantity
	}

	return requests, limits, nil
}

// convertKusionProbeToV1Probe converts Kusion Probe to Kubernetes Probe types.
func convertKusionProbeToV1Probe(p *v1.Probe) (*corev1.Probe, error) {
	result := &corev1.Probe{
		InitialDelaySeconds: p.InitialDelaySeconds,
		TimeoutSeconds:      p.TimeoutSeconds,
		PeriodSeconds:       p.PeriodSeconds,
		SuccessThreshold:    p.SuccessThreshold,
		FailureThreshold:    p.FailureThreshold,
	}
	probeHandler := p.ProbeHandler
	switch probeHandler.Type {
	case "Http":
		action, err := httpGetAction(probeHandler.HTTPGetAction.URL, probeHandler.Headers)
		if err != nil {
			return nil, err
		}
		result.HTTPGet = action
	case "Exec":
		result.Exec = &corev1.ExecAction{Command: probeHandler.Command}
	case "Tcp":
		action, err := tcpSocketAction(probeHandler.TCPSocketAction.URL)
		if err != nil {
			return nil, err
		}
		result.TCPSocket = action
	}
	return result, nil
}

// convertKusionLifecycleToV1Lifecycle converts Kusion Lifecycle to Kubernetes Lifecycle types.
func convertKusionLifecycleToV1Lifecycle(l *v1.Lifecycle) (*corev1.Lifecycle, error) {
	result := &corev1.Lifecycle{}
	if l.PreStop != nil {
		preStop, err := lifecycleHandler(l.PreStop)
		if err != nil {
			return nil, err
		}
		result.PreStop = preStop
	}
	if l.PostStart != nil {
		postStart, err := lifecycleHandler(l.PostStart)
		if err != nil {
			return nil, err
		}
		result.PostStart = postStart
	}
	return result, nil
}

func lifecycleHandler(in *v1.LifecycleHandler) (*corev1.LifecycleHandler, error) {
	result := &corev1.LifecycleHandler{}
	switch in.Type {
	case "Http":
		action, err := httpGetAction(in.HTTPGetAction.URL, in.Headers)
		if err != nil {
			return nil, err
		}
		result.HTTPGet = action
	case "Exec":
		result.Exec = &corev1.ExecAction{Command: in.Command}
	}
	return result, nil
}

func httpGetAction(urlstr string, headers map[string]string) (*corev1.HTTPGetAction, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	httpHeaders := make([]corev1.HTTPHeader, 0, len(headers))
	for k, v := range headers {
		httpHeaders = append(httpHeaders, corev1.HTTPHeader{
			Name:  k,
			Value: v,
		})
	}

	host := u.Hostname()
	if host == "localhost" || host == "127.0.0.1" {
		host = ""
	}

	return &corev1.HTTPGetAction{
		Path:        u.Path,
		Port:        intstr.Parse(u.Port()),
		Host:        host,
		Scheme:      corev1.URIScheme(strings.ToUpper(u.Scheme)),
		HTTPHeaders: httpHeaders,
	}, nil
}

func tcpSocketAction(urlstr string) (*corev1.TCPSocketAction, error) {
	host, port, err := net.ParseHostPort(urlstr)
	if err != nil {
		return nil, err
	}

	return &corev1.TCPSocketAction{
		Port: intstr.Parse(port),
		Host: host,
	}, nil
}

// handleFileCreation handles the creation of the files declared in container.Files
// and returns the generated ConfigMap, Volume and VolumeMount.
func handleFileCreation(c v1.Container, uniqueAppName, containerName string) (
	volumes []corev1.Volume,
	volumeMounts []corev1.VolumeMount,
	configMaps []corev1.ConfigMap,
	err error,
) {
	var idx int
	err = modules.ForeachOrdered(c.Files, func(k string, v v1.FileSpec) error {
		// The declared file path needs to include the file name.
		if filepath.Base(k) == "." || filepath.Base(k) == "/" {
			return fmt.Errorf("the declared file path needs to include the file name")
		}

		// Specify the name of the configMap and volume to be created.
		configMapName := uniqueAppName + "-" + containerName + "-" + strconv.Itoa(idx)
		idx++

		// Change the mode attribute from string into int32.
		var modeInt32 int32
		if modeInt64, err2 := strconv.ParseInt(v.Mode, 0, 64); err2 != nil {
			return err2
		} else {
			modeInt32 = int32(modeInt64)
		}

		if v.ContentFrom != "" {
			sec, ok, parseErr := parseSecretReference(v.ContentFrom)
			if parseErr != nil || !ok {
				return fmt.Errorf("invalid content from str")
			}

			volumes = append(volumes, corev1.Volume{
				Name: sec.Name,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  sec.Name,
						DefaultMode: &modeInt32,
					},
				},
			})

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      sec.Name,
				MountPath: filepath.Join("/", k),
				SubPath:   sec.Key,
			})
		} else if v.Content != "" {
			// Create the file content with configMap.
			data := make(map[string]string)
			data[filepath.Base(k)] = v.Content

			configMaps = append(configMaps, corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: corev1.SchemeGroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: configMapName,
				},
				Data: data,
			})

			volumes = append(volumes, corev1.Volume{
				Name: configMapName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName,
						},
						DefaultMode: &modeInt32,
					},
				},
			})

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      configMapName,
				MountPath: filepath.Dir(k),
			})
		}
		return nil
	})
	return
}

// handleDirCreation handles the creation of folder declared in container.Dirs and returns
// the generated Volume and VolumeMount.
func handleDirCreation(c v1.Container) (volumes []corev1.Volume, volumeMounts []corev1.VolumeMount, err error) {
	err = modules.ForeachOrdered(c.Dirs, func(mountPath string, v string) error {
		sec, ok, parseErr := parseSecretReference(v)
		if parseErr != nil || !ok {
			return fmt.Errorf("invalid dir configuration")
		}

		volumes = append(volumes, corev1.Volume{
			Name: sec.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: sec.Name,
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      sec.Name,
			MountPath: filepath.Join("/", mountPath),
		})
		return nil
	})
	return
}

// completeBaseWorkload uses config from workspace to complete the Workload base config.
func completeBaseWorkload(base *v1.Base, config v1.GenericConfig) error {
	replicas, err := workspace.GetInt32PointerFromGenericConfig(config, v1.FieldReplicas)
	if err != nil {
		return err
	}

	// override the base replicas with the value from workspace if it is null
	if base.Replicas == nil {
		base.Replicas = replicas
	}
	labels, err := workspace.GetStringMapFromGenericConfig(config, v1.FieldLabels)
	if err != nil {
		return err
	}
	if labels != nil {
		if err = mergo.Merge(&base.Labels, labels); err != nil {
			return err
		}
	}
	annotations, err := workspace.GetStringMapFromGenericConfig(config, v1.FieldAnnotations)
	if err != nil {
		return err
	}
	if annotations != nil {
		if err = mergo.Merge(&base.Annotations, annotations); err != nil {
			return err
		}
	}
	return nil
}

type secretReference struct {
	Name string
	Key  string
}

// parseSecretReference takes secret reference string as parameter and returns secretReference obj.
// Parameter `ref` is expected in following format: secret://sec-name/key, if the provided ref str
// is not in valid format, this function will return false or err.
func parseSecretReference(ref string) (result secretReference, _ bool, _ error) {
	if strings.HasPrefix(ref, "${secret://") && strings.HasSuffix(ref, "}") {
		ref = ref[2 : len(ref)-1]
	}

	if !strings.HasPrefix(ref, "secret://") {
		return result, false, nil
	}

	u, err := url.Parse(ref)
	if err != nil {
		return result, false, err
	}

	result.Name = u.Host
	result.Key, _, _ = strings.Cut(strings.TrimPrefix(u.Path, "/"), "/")

	return result, true, nil
}
