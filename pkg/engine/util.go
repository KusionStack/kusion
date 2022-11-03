package engine

const Separator = ":"

func BuildIDForKubernetes(apiVersion, kind, namespace, name string) string {
	key := apiVersion + Separator + kind + Separator
	if namespace != "" {
		key += namespace + Separator
	}
	return key + name
}
