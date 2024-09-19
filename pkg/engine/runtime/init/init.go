package init

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	apiv1 "kusionstack.io/kusion/pkg/apis/api.kusion.io/v1"
	v1 "kusionstack.io/kusion/pkg/apis/status/v1"
	"kusionstack.io/kusion/pkg/engine/operation/graph"
	"kusionstack.io/kusion/pkg/engine/runtime"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes"
	"kusionstack.io/kusion/pkg/engine/runtime/kubernetes/kubeops"
	"kusionstack.io/kusion/pkg/engine/runtime/terraform"
	"kusionstack.io/kusion/pkg/secrets"
	"kusionstack.io/kusion/pkg/workspace"
)

var SupportRuntimes = map[apiv1.Type]InitFn{
	runtime.Kubernetes: kubernetes.NewKubernetesRuntime,
	runtime.Terraform:  terraform.NewTerraformRuntime,
}

var contextKeys = []string{
	kubeops.KubeConfigContentKey,
	apiv1.EnvAwsAccessKeyID,
	apiv1.EnvAwsSecretAccessKey,
	apiv1.EnvAlicloudAccessKey,
	apiv1.EnvAlicloudSecretKey,
}

// InitFn runtime init func
type InitFn func(spec apiv1.Spec) (runtime.Runtime, error)

func Runtimes(spec apiv1.Spec, state apiv1.State) (map[apiv1.Type]runtime.Runtime, v1.Status) {
	// Parse the secret ref in the Context of Spec.
	if err := parseContextSecretRef(&spec); err != nil {
		return nil, v1.NewErrorStatus(err)
	}
	resources := spec.Resources
	resources = append(resources, state.Resources...)
	runtimesMap := map[apiv1.Type]runtime.Runtime{}
	if resources == nil {
		return runtimesMap, nil
	}
	if errStatus := validResources(resources); errStatus != nil {
		return nil, errStatus
	}

	for _, resource := range resources {
		rt := resource.Type
		if runtimesMap[rt] == nil {
			r, err := SupportRuntimes[rt](spec)
			if err != nil {
				return nil, v1.NewErrorStatus(fmt.Errorf("init %s runtime failed. %w", rt, err))
			}
			runtimesMap[rt] = r
		}
	}
	return runtimesMap, nil
}

func validResources(resources apiv1.Resources) v1.Status {
	var kubeConfig string
	for _, resource := range resources {
		rt := resource.Type
		if rt == "" {
			return v1.NewErrorStatusWithCode(v1.IllegalManifest, fmt.Errorf("no resource type in resource: %v", resource.ID))
		}
		if SupportRuntimes[rt] == nil {
			return v1.NewErrorStatusWithCode(v1.IllegalManifest, fmt.Errorf("unknown resource type: %s. Currently supported resource types are: %v",
				rt, reflect.ValueOf(SupportRuntimes).MapKeys()))
		}
		if rt == apiv1.Kubernetes {
			config := kubeops.GetKubeConfig(&resource)
			if kubeConfig != "" && kubeConfig != config {
				return v1.NewErrorStatusWithCode(v1.IllegalManifest, fmt.Errorf("different kubeConfig in different resources"))
			}
			if kubeConfig == "" {
				kubeConfig = config
			}
		}
	}
	return nil
}

// parseContextSecretRef parses the external secret ref of the credentials
// in the Context of Spec.
func parseContextSecretRef(spec *apiv1.Spec) error {
	// Copy the Context of Spec.
	parsedContext := apiv1.GenericConfig{}
	for k, v := range spec.Context {
		parsedContext[k] = v
	}

	// Retrieve the context with the specified keys from spec and parse the external secret ref.
	for _, key := range contextKeys {
		contextStr, err := workspace.GetStringFromGenericConfig(spec.Context, key)
		if err != nil {
			return err
		}

		if contextStr != "" {
			// Replace the secret store ref.
			if strings.HasPrefix(contextStr, graph.SecretRefPrefix) {
				externalSecretRef, err := graph.ParseExternalSecretDataRef(contextStr)
				if err != nil {
					return err
				}

				provider, exist := secrets.GetProvider(spec.SecretStore.Provider)
				if !exist {
					return errors.New("no matched secret store found, please check workspace yaml")
				}
				secretStore, err := provider.NewSecretStore(spec.SecretStore)
				if err != nil {
					return err
				}
				secretData, err := secretStore.GetSecret(context.Background(), *externalSecretRef)
				if err != nil {
					return err
				}

				parsedContext[key] = string(secretData)
			}
		}
	}

	// Reset the Context with the parsed values.
	spec.Context = parsedContext

	return nil
}
