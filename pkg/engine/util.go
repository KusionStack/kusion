package engine

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

const Separator = ":"

func BuildID(apiVersion, kind, namespace, name string) string {
	key := apiVersion + Separator + kind + Separator
	if namespace != "" {
		key += namespace + Separator
	}
	return key + name
}

func BuildIDForKubernetes(o *unstructured.Unstructured) string {
	return BuildID(o.GetAPIVersion(), o.GetKind(), o.GetNamespace(), o.GetName())
}
