package workload

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/apis/core/v1/workload"
	"kusionstack.io/kusion/pkg/modules"
)

type jobGenerator struct {
	project   string
	stack     string
	appName   string
	job       *workload.Job
	jobConfig apiv1.GenericConfig
	namespace string
}

func NewJobGenerator(generator *Generator) (modules.Generator, error) {
	return &jobGenerator{
		project:   generator.Project,
		stack:     generator.Stack,
		appName:   generator.App,
		job:       generator.Workload.Job,
		jobConfig: generator.PlatformConfigs[workload.ModuleJob],
		namespace: generator.Namespace,
	}, nil
}

func NewJobGeneratorFunc(generator *Generator) modules.NewGeneratorFunc {
	return func() (modules.Generator, error) {
		return NewJobGenerator(generator)
	}
}

func (g *jobGenerator) Generate(spec *apiv1.Intent) error {
	job := g.job
	if job == nil {
		return nil
	}

	if spec.Resources == nil {
		spec.Resources = make(apiv1.Resources, 0)
	}

	if err := completeBaseWorkload(&g.job.Base, g.jobConfig); err != nil {
		return fmt.Errorf("complete job input by workspace config failed, %w", err)
	}

	uniqueAppName := modules.UniqueAppName(g.project, g.stack, g.appName)

	meta := metav1.ObjectMeta{
		Namespace: g.namespace,
		Name:      uniqueAppName,
		Labels: modules.MergeMaps(
			modules.UniqueAppLabels(g.project, g.appName),
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
		cmObj.Namespace = g.namespace
		if err = modules.AppendToIntent(
			apiv1.Kubernetes,
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
					modules.UniqueAppLabels(g.project, g.appName),
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
		return modules.AppendToIntent(apiv1.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
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
	return modules.AppendToIntent(apiv1.Kubernetes, modules.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
}
