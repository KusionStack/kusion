package config

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"

	"kusionstack.io/kusion/pkg/apis/stack"
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

// 1. If $KUBECONFIG environment variable is set, then it is used.
// 2. If not, and the `kubeConfig` in stack configuration is set, then it is used.
// 3. Otherwise, ${HOME}/.kube/config is used.
func GetKubeConfig(stack *stack.Stack) string {
	if kubeConfigFile := os.Getenv(RecommendedConfigPathEnvVar); kubeConfigFile != "" {
		return kubeConfigFile
	} else if stack.KubeConfig != "" {
		kubeConfigFile, _ := filepath.Abs(stack.KubeConfig)
		if kubeConfigFile != "" {
			return kubeConfigFile
		}
	}
	return RecommendedKubeConfigFile
}
