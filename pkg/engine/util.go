package engine

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// ContextKey is used to represent the key associated with the information
// injected into the function context.
type ContextKey string

const (
	// WatchChannel is used to inject a channel into the runtime operation to
	// assist in watching the changes of the resources during the operation process.
	WatchChannel = ContextKey("WatchChannel")
)

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
