package workload

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/apis/project"
	"kusionstack.io/kusion/pkg/apis/stack"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

type jobGenerator struct {
	project *project.Project
	stack   *stack.Stack
	appName string
	job     *workload.Job
}

func NewJobGenerator(
	project *project.Project,
	stack *stack.Stack,
	appName string,
	job *workload.Job,
) (modules.Generator, error) {
	return &jobGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		job:     job,
	}, nil
}

func NewJobGeneratorFunc(
	project *project.Project,
	stack *stack.Stack,
	appName string,
	job *workload.Job,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewJobGenerator(project, stack, appName, job)
	}
}

func (g *jobGenerator) Generate(spec *intent.Intent) error {
	job := g.job
	if job == nil {
		return nil
	}

	if spec.Resources == nil {
		spec.Resources = make(intent.Resources, 0)
	}

	uniqueAppName := modules.UniqueAppName(g.project.Name, g.stack.Name, g.appName)

	meta := metav1.ObjectMeta{
		Namespace: g.project.Name,
		Name:      uniqueAppName,
		Labels: modules.MergeMaps(
			modules.UniqueAppLabels(g.project.Name, g.appName),
			g.job.Labels,
		),
		Annotations: modules.MergeMaps(
			g.job.Annotations,
		),
	}

	containers, volumes, configMaps, err := toOrderedContainers(job.Containers, uniqueAppName)
	if err != nil {
		return err
	}

	for _, cm := range configMaps {
		cmObj := cm
		cmObj.Namespace = g.project.Name
		if err = modules.AppendToSpec(
			intent.Kubernetes,
			modules.KubernetesResourceID(cmObj.TypeMeta, cmObj.ObjectMeta),
			spec,
			&cmObj,
		); err != nil {
			return err
		}
	}

	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: modules.MergeMaps(
					modules.UniqueAppLabels(g.project.Name, g.appName),
					g.job.Labels,
				),
				Annotations: modules.MergeMaps(
					g.job.Annotations,
				),
			},
			Spec: corev1.PodSpec{
				Containers: containers,
				Volumes:    volumes,
			},
		},
	}

	if job.Schedule == "" {
		resource := &batchv1.Job{
			ObjectMeta: meta,
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: batchv1.SchemeGroupVersion.String(),
			},
			Spec: jobSpec,
		}
		return modules.AppendToSpec(intent.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
	}

	resource := &batchv1.CronJob{
		ObjectMeta: meta,
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: batchv1.SchemeGroupVersion.String(),
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: jobSpec,
			},
			Schedule: job.Schedule,
		},
	}
	return modules.AppendToSpec(intent.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
}
