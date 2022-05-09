package config

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/util/homedir"
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

// 1. If $KUBECONFIG environment variable is set, then it is used it.
// 2. Otherwise, ${HOME}/.kube/config is used.
func GetKubeConfig() string {
	if kubeConfigFile := os.Getenv(RecommendedConfigPathEnvVar); kubeConfigFile != "" {
		return kubeConfigFile
	} else {
		return RecommendedKubeConfigFile
	}
}
