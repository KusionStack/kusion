package e2e

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kusionGenerateCmd = "kusion generate"
	kusionPreviewCmd  = "kusion preview -d=false"
	kusionApplyCmd    = "kusion apply --watch=false -y=true"
	kusionDestroyCmd  = "kusion destroy -y=true"
	kusionVersionCmd  = "kusion version"
)

var _ = ginkgo.Describe("Kusion Configuration Commands", func() {
	ginkgo.Context("kusion generate testing", func() {
		ginkgo.It("kusion generate", func() {
			// kusion build testing
			path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
			if runtime.GOOS == "windows" {
				kusionGenerateCmd = "kusion.exe generate"
			}
			output, err := ExecKusionWithWorkDir(kusionGenerateCmd, path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("Generating Spec"))
		})
	})
})

var _ = ginkgo.Describe("kusion Runtime Commands", func() {
	ginkgo.It("kusion preview", func() {
		path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
		if runtime.GOOS == "windows" {
			kusionPreviewCmd = "kusion.exe preview -d=false"
		}
		_, err := ExecKusionWithWorkDir(kusionPreviewCmd, path)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})

	ginkgo.It("apply and destroy", func() {
		ginkgo.By("kusion apply", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
			if runtime.GOOS == "windows" {
				kusionApplyCmd = "kusion.exe apply --watch=false -y=true"
			}
			_, err := ExecKusionWithWorkDir(kusionApplyCmd, path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("kusion apply", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "quickstart", "default")
			if runtime.GOOS == "windows" {
				kusionApplyCmd = "kusion.exe apply --watch=false -y=true"
			}
			_, err := ExecKusionWithWorkDir(kusionApplyCmd, path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("wait service-multi-stack deploy", func() {
			homedir := os.Getenv("HOME")
			configPath := filepath.Join(homedir, ".kube", "config")
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.AppsV1().Deployments("service-multi-stack").Get(context.TODO(), "service-multi-stack-dev-echoserver", metav1.GetOptions{})
				return err == nil
			}, 900*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})

		ginkgo.By("wait quickstart deploy", func() {
			homedir := os.Getenv("HOME")
			configPath := filepath.Join(homedir, ".kube", "config")
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.AppsV1().Deployments("quickstart").Get(context.TODO(), "quickstart-default-quickstart", metav1.GetOptions{})
				return err == nil
			}, 900*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})

		ginkgo.By("kusion destroy", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
			if runtime.GOOS == "windows" {
				kusionDestroyCmd = "kusion.exe destroy -y=true"
			}
			_, err := ExecKusionWithWorkDir(kusionDestroyCmd, path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("kusion destroy", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "quickstart", "default")
			if runtime.GOOS == "windows" {
				kusionDestroyCmd = "kusion.exe destroy -y=true"
			}
			_, err := ExecKusionWithWorkDir(kusionDestroyCmd, path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("wait service-multi-stack destroy", func() {
			homedir := os.Getenv("HOME")
			configPath := filepath.Join(homedir, ".kube", "config")
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.CoreV1().Namespaces().Get(context.TODO(), "service-multi-stack", metav1.GetOptions{})
				return errors.IsNotFound(err)
			}, 900*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})

		ginkgo.By("wait service-multi-stack destroy", func() {
			homedir := os.Getenv("HOME")
			configPath := filepath.Join(homedir, ".kube", "config")
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.CoreV1().Namespaces().Get(context.TODO(), "quickstart", metav1.GetOptions{})
				return errors.IsNotFound(err)
			}, 900*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})
	})
})

var _ = ginkgo.Describe("Kusion Other Commands", func() {
	ginkgo.Context("kusion version testing", func() {
		ginkgo.It("kusion version", func() {
			if runtime.GOOS == "windows" {
				kusionVersionCmd = "kusion.exe version"
			}
			output, err := ExecKusion(kusionVersionCmd)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("releaseVersion"))
		})
	})
})
