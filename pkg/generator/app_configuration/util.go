package app_configuration

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildK8sResourceId(typeMeta metav1.TypeMeta, objectMeta metav1.ObjectMeta) string {
	// resource id example: apps/v1:Deployment:code-city:code-citydev
	id := typeMeta.APIVersion + ":" + typeMeta.Kind + ":"
	if objectMeta.Namespace != "" {
		id = id + objectMeta.Namespace + ":"
	}
	id = id + objectMeta.Name
	return id
}
