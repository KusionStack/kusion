package generators

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kusionstack.io/kusion/pkg/models"
	"kusionstack.io/kusion/pkg/models/appconfiguration/component"
)

type jobGenerator struct {
	projectName string
	compName    string
	comp        *component.Component
}

func NewJobGenerator(projectName, compName string, comp *component.Component) (Generator, error) {
	return &jobGenerator{
		projectName: projectName,
		compName:    compName,
		comp:        comp,
	}, nil
}

func NewJobGeneratorFunc(projectName, compName string, comp *component.Component) NewGeneratorFunc {
	return func() (Generator, error) {
		return NewJobGenerator(projectName, compName, comp)
	}
}

func (g *jobGenerator) Generate(spec *models.Spec) error {
	if g.comp.WorkloadType != component.WorkloadTypeJob {
		return nil
	}

	if spec.Resources == nil {
		spec.Resources = make(models.Resources, 0)
	}

	meta := metav1.ObjectMeta{
		Namespace:   g.projectName,
		Name:        uniqueComponentName(g.projectName, g.compName),
		Labels:      g.comp.Labels,
		Annotations: g.comp.Annotations,
	}

	containers, err := toOrderedContainers(g.comp.Containers)
	if err != nil {
		return err
	}
	jobSpec := batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: uniqueComponentLabels(g.projectName, g.compName),
			},
			Spec: corev1.PodSpec{
				Containers: containers,
			},
		},
	}

	if g.comp.Schedule == "" {
		resource := &batchv1.Job{
			ObjectMeta: meta,
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "batch/v1",
			},
			Spec: jobSpec,
		}
		return appendToSpec(
			kubernetesResourceID(resource.TypeMeta, resource.ObjectMeta),
			resource,
			spec,
		)
	}

	resource := &batchv1.CronJob{
		ObjectMeta: meta,
		TypeMeta: metav1.TypeMeta{
			Kind:       "CronJob",
			APIVersion: "batch/v1",
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: jobSpec,
			},
			Schedule: g.comp.Schedule,
		},
	}
	return appendToSpec(
		kubernetesResourceID(resource.TypeMeta, resource.ObjectMeta),
		resource,
		spec,
	)
}
