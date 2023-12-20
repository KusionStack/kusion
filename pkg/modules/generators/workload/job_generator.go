package workload

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/intent"
	"kusionstack.io/kusion/pkg/modules"
	"kusionstack.io/kusion/pkg/modules/inputs/workload"
)

type jobGenerator struct {
	project   *apiv1.Project
	stack     *apiv1.Stack
	appName   string
	job       *workload.Job
	jobConfig apiv1.GenericConfig
}

func NewJobGenerator(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	job *workload.Job,
	jobConfig apiv1.GenericConfig,
) (modules.Generator, error) {
	return &jobGenerator{
		project:   project,
		stack:     stack,
		appName:   appName,
		job:       job,
		jobConfig: jobConfig,
	}, nil
}

func NewJobGeneratorFunc(
	project *apiv1.Project,
	stack *apiv1.Stack,
	appName string,
	job *workload.Job,
	jobConfig apiv1.GenericConfig,
) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewJobGenerator(project, stack, appName, job, jobConfig)
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

	if err := completeBaseWorkload(&g.job.Base, g.jobConfig); err != nil {
		return fmt.Errorf("complete job input by workspace config failed, %w", err)
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
		if err = modules.AppendToIntent(
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
				Containers:    containers,
				RestartPolicy: corev1.RestartPolicyNever,
				Volumes:       volumes,
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
		return modules.AppendToIntent(intent.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
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
	return modules.AppendToIntent(intent.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
}
