package workload

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/generators/workload/secret"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
	"kusionstack.io/kusion/pkg/modules/inputs/workload/container"
	"kusionstack.io/kusion/pkg/util/net"
	"kusionstack.io/kusion/pkg/workspace"
)

type workloadGenerator struct {
	project       *apiv1.Project
	stack         *apiv1.Stack
	appName       string
	workload      *workload.Workload
	moduleConfigs map[string]apiv1.GenericConfig
	namespace     string

	// for internal generator
	context modules.GeneratorContext
}

func NewWorkloadGenerator(ctx modules.GeneratorContext) (modules.Generator, error) {
	if len(ctx.Project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &workloadGenerator{
		project:       ctx.Project,
		stack:         ctx.Stack,
		appName:       ctx.Application.Name,
		workload:      ctx.Application.Workload,
		moduleConfigs: ctx.ModuleInputs,
		namespace:     ctx.Namespace,
		context:       ctx,
	}, nil
}

func NewWorkloadGeneratorFunc(ctx modules.GeneratorContext) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewWorkloadGenerator(ctx)
	}
}

func (g *workloadGenerator) Generate(spec *apiv1.Intent) error {
	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	if g.workload != nil {
		var gfs []modules.NewGeneratorFunc

		switch g.workload.Header.Type {
		case workload.TypeService:
			gfs = append(gfs,
				NewWorkloadServiceGeneratorFunc(g.context),
				secret.NewSecretGeneratorFunc(g.context))
		case workload.TypeJob:
			gfs = append(gfs,
				NewJobGeneratorFunc(g.context),
				secret.NewSecretGeneratorFunc(g.context))
		}

		if err := modules.CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}

	return nil
}

func toOrderedContainers(
	appContainers map[string]container.Container,
	uniqueAppName string,
) ([]corev1.Container, []corev1.Volume, []corev1.ConfigMap, error) {
	// Create a slice of containers based on the app's
	// containers.
	var containers []corev1.Container

	// Create a slice of volumes and configMaps based on the containers' files to be created.
	var volumes []corev1.Volume
	var volumeMounts []corev1.VolumeMount
	var configMaps []corev1.ConfigMap

	if err := modules.ForeachOrdered(appContainers, func(containerName string, c container.Container) error {
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
	err = modules.ForeachOrdered(c.Files, func(k string, v container.FileSpec) error {
		// for k, v := range c.Files {
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
		return nil
	})
	return
}

// completeBaseWorkload uses config from workspace to complete the workload base config.
func completeBaseWorkload(base *workload.Base, config apiv1.GenericConfig) error {
	replicas, err := workspace.GetIntFromGenericConfig(config, workload.FieldReplicas)
	if err != nil {
		return err
	}
	if replicas == 0 {
		replicas = workload.DefaultReplicas
	}
	if base.Replicas == 0 {
		base.Replicas = replicas
	}
	labels, err := workspace.GetStringMapFromGenericConfig(config, workload.FieldLabels)
	if err != nil {
		return err
	}
	if labels != nil {
		if err = mergo.Merge(&base.Labels, labels); err != nil {
			return err
		}
	}
	annotations, err := workspace.GetStringMapFromGenericConfig(config, workload.FieldAnnotations)
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
