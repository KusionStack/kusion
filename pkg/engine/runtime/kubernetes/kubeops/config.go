package kubeops

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"

	"kusionstack.io/kusion/pkg/apis/intent"
)

const (
	RecommendedConfigPathEnvVar   = "KUBECONFIG"
	RecommendedHomeDir            = ".kube"
	RecommendedKubeConfigFileName = "config"
)

var (
	RecommendedConfigDir      = filepath.Join(homedir.HomeDir(), RecommendedHomeDir)
	RecommendedKubeConfigFile = filepath.Join(RecommendedConfigDir, RecommendedKubeConfigFileName)
)

// GetKubeConfig gets kubeConfig in the following order:
// 1. If $KUBECONFIG environment variable is set, then it is used.
// 2. If not, and the `kubeConfig` in resource extensions is set, then it is used.
// 3. Otherwise, ${HOME}/.kube/config is used.
func GetKubeConfig(resource *intent.Resource) string {
	if kubeConfigFile := os.Getenv(RecommendedConfigPathEnvVar); kubeConfigFile != "" {
		return kubeConfigFile
	}
	if resource != nil {
		kubeConfig, ok := resource.Extensions[intent.ResourceExtensionKubeConfig].(string)
		if ok && kubeConfig != "" {
			kubeConfigFile, _ := filepath.Abs(kubeConfig)
			if kubeConfigFile != "" {
				return kubeConfigFile
			}
		}
	}
	return RecommendedKubeConfigFile
}
