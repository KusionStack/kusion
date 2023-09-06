package workload

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/monitoring"
	"kusionstack.io/kusion/pkg/models/appconfiguration/trait"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/net"
)

type workloadGenerator struct {
	project    *projectstack.Project
	stack      *projectstack.Stack
	appName    string
	workload   *workload.Workload
	monitoring *monitoring.Monitor
	opsRule    *trait.OpsRule
}

func NewWorkloadGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workload *workload.Workload,
	monitoring *monitoring.Monitor,
	opsRule *trait.OpsRule,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &workloadGenerator{
		project:    project,
		stack:      stack,
		appName:    appName,
		workload:   workload,
		monitoring: monitoring,
		opsRule:    opsRule,
	}, nil
}

func NewWorkloadGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workload *workload.Workload,
	monitoring *monitoring.Monitor,
	opsRule *trait.OpsRule,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewWorkloadGenerator(project, stack, appName, workload, monitoring, opsRule)
	}
}

func (g *workloadGenerator) Generate(spec *models.Spec) error {
	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	if g.workload != nil {
		var gfs []appconfiguration.NewGeneratorFunc

		switch g.workload.Header.Type {
		case workload.TypeService:
			gfs = append(gfs, NewWorkloadServiceGeneratorFunc(g.project, g.stack, g.appName, g.workload.Service, g.monitoring, g.opsRule))
		case workload.TypeJob:
			gfs = append(gfs, NewJobGeneratorFunc(g.project, g.stack, g.appName, g.workload.Job))
		}

		if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}

	return nil
}

func toOrderedContainers(appContainers map[string]container.Container, uniqueAppName string) ([]corev1.Container, []corev1.Volume, []corev1.ConfigMap, error) {
	// Create a slice of containers based on the app's
	// containers.
	var containers []corev1.Container

	// Create a slice of volumes and configMaps based on the containers' files to be created.
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var configMaps []corev1.ConfigMap

	if err := appconfiguration.ForeachOrdered(appContainers, func(containerName string, c container.Container) error {
		// Create a slice of env vars based on the container's env vars.
		var envs []corev1.EnvVar
		for k, v := range c.Env {
			envs = append(envs, *MagicEnvVar(k, v))
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

		// Append the container object to the containers slice.
		containers = append(containers, ctn)
		return nil
	}); err != nil {
		return nil, nil, nil, err
	}
	return containers, volumes, configMaps, nil
}

// updateContainer updates corev1.Container with passed parameters.
func updateContainer(in *container.Container, out *corev1.Container) error {
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
func convertKusionProbeToV1Probe(p *container.Probe) (*corev1.Probe, error) {
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
func convertKusionLifecycleToV1Lifecycle(l *container.Lifecycle) (*corev1.Lifecycle, error) {
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

func lifecycleHandler(in *container.LifecycleHandler) (*corev1.LifecycleHandler, error) {
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

// handleFileCreation handles the creation of the files declared in container.File
// and returns the generated ConfigMap, Volume and VolumeMount.
func handleFileCreation(c container.Container, uniqueAppName, containerName string) (
	volumes []corev1.Volume,
	volumeMounts []corev1.VolumeMount,
	configMaps []corev1.ConfigMap,
	err error,
) {
	var idx int
	for k, v := range c.Files {
		// The declared file path needs to include the file name.
		if filepath.Base(k) == "." || filepath.Base(k) == "/" {
			err = fmt.Errorf("the declared file path needs to include the file name")
			return
		}

		// Specify the name of the configMap and volume to be created.
		configMapName := uniqueAppName + "-" + containerName + "-" + strconv.Itoa(idx)
		idx++

		// Change the mode attribute from string into int32.
		var modeInt32 int32
		var modeInt64 int64
		if modeInt64, err = strconv.ParseInt(v.Mode, 0, 64); err != nil {
			return
		} else {
			modeInt32 = int32(modeInt64)
		}

		if v.ContentFrom != "" {
			// TODO: support the creation of the file content from a reference source.
			panic("not supported the creation the file content from a reference source")
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
	}
	return
}
