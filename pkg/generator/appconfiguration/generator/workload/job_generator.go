package workload

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kusionstack.io/kusion/pkg/generator/appconfiguration"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/workload"
	"kusionstack.io/kusion/pkg/projectstack"
)

type jobGenerator struct {
	project *projectstack.Project
	stack   *projectstack.Stack
	appName string
	job     *workload.Job
}

func NewJobGenerator(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	job *workload.Job,
) (appconfiguration.Generator, error) {
	return &jobGenerator{
		project: project,
		stack:   stack,
		appName: appName,
		job:     job,
	}, nil
}

func NewJobGeneratorFunc(
	project *projectstack.Project,
	stack *projectstack.Stack,
	appName string,
	job *workload.Job,
) appconfiguration.NewGeneratorFunc {
	return func() (appconfiguration.Generator, error) {
		return NewJobGenerator(project, stack, appName, job)
	}
}

func (g *jobGenerator) Generate(spec *models.Spec) error {
	job := g.job
	if job == nil {
		return nil
	}

	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	uniqueAppName := appconfiguration.UniqueAppName(g.project.Name, g.stack.Name, g.appName)

	meta := metav1.ObjectMeta{
		Namespace: g.project.Name,
		Name:      uniqueAppName,
		Labels: appconfiguration.MergeMaps(
			appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
			g.job.Labels,
		),
		Annotations: appconfiguration.MergeMaps(
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
		if err = appconfiguration.AppendToSpec(
			models.Kubernetes,
			appconfiguration.KubernetesResourceID(cmObj.TypeMeta, cmObj.ObjectMeta),
			spec,
			&cmObj,
		); err != nil {
			return err
		}
	}

	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: appconfiguration.MergeMaps(
					appconfiguration.UniqueAppLabels(g.project.Name, g.appName),
					g.job.Labels,
				),
				Annotations: appconfiguration.MergeMaps(
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
		return appconfiguration.AppendToSpec(models.Kubernetes, appconfiguration.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
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
	return appconfiguration.AppendToSpec(models.Kubernetes, appconfiguration.KubernetesResourceID(resource.TypeMeta, resource.ObjectMeta), spec, resource)
}
