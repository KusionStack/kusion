package diff

import (
	"encoding/json"

	"kusionstack.io/kusion/pkg/log"

	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type normalizerPatch struct {
	groupKind schema.GroupKind
	namespace string
	name      string
	patch     jsonpatch.Patch
}

type ignoreNormalizer struct {
	patches []normalizerPatch
}

// Normalize removes fields from supplied resource using json paths from matching items of specified resources ignored differences list
func (n *ignoreNormalizer) Normalize(un *unstructured.Unstructured) error {
	matched := make([]normalizerPatch, 0)
	for _, patch := range n.patches {
		groupKind := un.GroupVersionKind().GroupKind()

		if patch.groupKind.Group == groupKind.Group &&
			patch.groupKind.Kind == groupKind.Kind &&
			(patch.name == "" || patch.name == un.GetName()) &&
			(patch.namespace == "" || patch.namespace == un.GetNamespace()) {
			matched = append(matched, patch)
		}
	}
	if len(matched) == 0 {
		return nil
	}

	docData, err := json.Marshal(un)
	if err != nil {
		return err
	}

	for _, patch := range matched {
		patchedData, err := patch.patch.Apply(docData)
		if err != nil {
			log.Debugf("Failed to apply normalization: %v", err)
			continue
		}
		docData = patchedData
	}

	return json.Unmarshal(docData, un)
}

// NewDefaultIgnoreNormalizer creates diff normalizer which removes ignored fields according to given json path
func NewDefaultIgnoreNormalizer(paths []string) (Normalizer, error) {
	patches := make([]normalizerPatch, 0)
	for _, path := range paths {
		patchData, err := json.Marshal([]map[string]string{{"op": "remove", "path": path}})
		if err != nil {
			return nil, err
		}
		patch, err := jsonpatch.DecodePatch(patchData)
		if err != nil {
			return nil, err
		}
		// Note: hardcode for now, should support customize configuration
		if path == "/metadata/annotations/helm.sh~1hook" || path == "/metadata/annotations/helm.sh~1hook-weight" {
			patches = append(patches, normalizerPatch{
				groupKind: schema.GroupKind{
					Group: "apiextensions.k8s.io",
					Kind:  "CustomResourceDefinition",
				},
				patch: patch,
			})
		}
	}
	return &ignoreNormalizer{
		patches: patches,
	}, nil
}
