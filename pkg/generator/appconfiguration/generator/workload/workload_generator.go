package workload

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload/container"
	"kusionstack.io/kusion/pkg/projectstack"
	"kusionstack.io/kusion/pkg/util/net"
)

type workloadGenerator struct {
	project  *projectstack.Project
	stack    *projectstack.Stack
	appName  string
	workload *workload.Workload
}

func NewWorkloadGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workload *workload.Workload,
) (appconfiguration.Generator, error) {
	if len(project.Name) == 0 {
		return nil, fmt.Errorf("project name must not be empty")
	}

	return &workloadGenerator{
		project:  project,
		stack:    stack,
		appName:  appName,
		workload: workload,
	}, nil
}

func NewWorkloadGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	workload *workload.Workload,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewWorkloadGenerator(project, stack, appName, workload)
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
			gfs = append(gfs, NewWorkloadServiceGeneratorFunc(g.project, g.stack, g.appName, g.workload.Service))
		case workload.TypeJob:
			gfs = append(gfs, NewJobGeneratorFunc(g.project, g.stack, g.appName, g.workload.Job))
		}

		if err := appconfiguration.CallGenerators(spec, gfs...); err != nil {
			return err
		}
	}

	return nil
}

func toOrderedContainers(appContainers map[string]container.Container) ([]v1.Container, error) {
	// Create a slice of containers based on the app's
	// containers.
	var containers []v1.Container
	if err := appconfiguration.ForeachOrdered(appContainers, func(containerName string, c container.Container) error {
		// Create a slice of env vars based on the container's env vars.
		var envs []v1.EnvVar
		for k, v := range c.Env {
			envs = append(envs, *MagicEnvVar(k, v))
		}
		resourceRequirements, err := handleResourceRequirementsV1(c.Resources)
		if err != nil {
			return err
		}

		ctn := v1.Container{
			Name:       containerName,
			Image:      c.Image,
			Command:    c.Command,
			Args:       c.Args,
			WorkingDir: c.WorkingDir,
			Env:        envs,
			Resources:  resourceRequirements,
		}
		err = updateContainer(&c, &ctn)
		if err != nil {
			return err
		}
		// Create a container object and append it to the containers slice.
		containers = append(containers, ctn)
		return nil
	}); err != nil {
		return nil, err
	}
	return containers, nil
}

// updateContainer updates v1.Container with passed parameters.
func updateContainer(in *container.Container, out *v1.Container) error {
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
func handleResourceRequirementsV1(resources map[string]string) (v1.ResourceRequirements, error) {
	result := v1.ResourceRequirements{}
	if resources == nil {
		return result, nil
	}
	for key, value := range resources {
		resourceName := v1.ResourceName(key)
		requests, limits, err := populateResourceLists(resourceName, value)
		if err != nil {
			return result, err
		}
		if requests != nil && result.Requests == nil {
			result.Requests = make(v1.ResourceList)
		}
		maps.Copy(result.Requests, requests)
		if limits != nil && result.Limits == nil {
			result.Limits = make(v1.ResourceList)
		}
		maps.Copy(result.Limits, limits)
	}
	return result, nil
}

// populateResourceLists takes strings of form <resourceName>=[<minValue>-]<maxValue> and
// returns request&limit ResourceList.
func populateResourceLists(name v1.ResourceName, spec string) (v1.ResourceList, v1.ResourceList, error) {
	requests := v1.ResourceList{}
	limits := v1.ResourceList{}

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
func convertKusionProbeToV1Probe(p *container.Probe) (*v1.Probe, error) {
	result := &v1.Probe{
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
		result.Exec = &v1.ExecAction{Command: probeHandler.Command}
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
func convertKusionLifecycleToV1Lifecycle(l *container.Lifecycle) (*v1.Lifecycle, error) {
	result := &v1.Lifecycle{}
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

func lifecycleHandler(in *container.LifecycleHandler) (*v1.LifecycleHandler, error) {
	result := &v1.LifecycleHandler{}
	switch in.Type {
	case "Http":
		action, err := httpGetAction(in.HTTPGetAction.URL, in.Headers)
		if err != nil {
			return nil, err
		}
		result.HTTPGet = action
	case "Exec":
		result.Exec = &v1.ExecAction{Command: in.Command}
	}
	return result, nil
}

func httpGetAction(urlstr string, headers map[string]string) (*v1.HTTPGetAction, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return nil, err
	}

	httpHeaders := make([]v1.HTTPHeader, 0, len(headers))
	for k, v := range headers {
		httpHeaders = append(httpHeaders, v1.HTTPHeader{
			Name:  k,
			Value: v,
		})
	}

	return &v1.HTTPGetAction{
		Path:        u.Path,
		Port:        intstr.FromString(u.Port()),
		Host:        u.Hostname(),
		Scheme:      v1.URIScheme(strings.ToUpper(u.Scheme)),
		HTTPHeaders: httpHeaders,
	}, nil
}

func tcpSocketAction(urlstr string) (*v1.TCPSocketAction, error) {
	host, port, err := net.ParseHostPort(urlstr)
	if err != nil {
		return nil, err
	}

	return &v1.TCPSocketAction{
		Port: intstr.FromString(port),
		Host: host,
	}, nil
}
