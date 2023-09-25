package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("Kusion Configuration Commands", func() {
	ginkgo.Context("kusion compile testing", func() {
		ginkgo.It("kusion compile", func() {
			// kusion compile testing
			path := filepath.Join(GetWorkDir(), "konfig", "example", "multi-stack", "dev")
			output, err := ExecKusionWithWorkDir("kusion compile", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("Generating Spec"))
		})
	})

	ginkgo.Context("kusion check testing", func() {
		ginkgo.It("kusion check", func() {
			// kusion check testing
			path := filepath.Join(GetWorkDir(), "konfig", "example", "multi-stack", "dev")
			output, err := ExecKusionWithWorkDir("kusion check", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("Generating Spec"))
		})
	})

	ginkgo.Context("kusion ls testing", func() {
		ginkgo.It("kusion ls", func() {
			path := filepath.Join(GetWorkDir(), "konfig")
			output, err := ExecKusionWithWorkDir("kusion ls --format=json", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("example", "multi-stack"))
		})
	})

	ginkgo.Context("kusion deps testing", func() {
		ginkgo.It("kusion deps", func() {
			// kusion deps testing
			path := filepath.Join(GetWorkDir(), "konfig")
			output, err := ExecKusionWithWorkDir("kusion deps --focus example/multi-stack/dev/main.k", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("catalog/models/schema/v1"))
		})
	})
})

var _ = ginkgo.Describe("kusion Runtime Commands", func() {
	ginkgo.It("kusion preview", func() {
		path := filepath.Join(GetWorkDir(), "konfig", "example", "multi-stack", "dev")
		_, err := ExecKusionWithWorkDir("kusion preview -d", path)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})

	ginkgo.It("kusion apply", func() {
		ginkgo.By("kusion apply", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "multi-stack", "dev")
			_, err := ExecKusionWithWorkDir("kusion apply -y=true", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("wait multi-stack deploy", func() {
			homedir := os.Getenv("HOME")
			configPath := fmt.Sprintf("%s/.kube/config", homedir)
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.AppsV1().Deployments("multi-stack").Get(context.TODO(), "multi-stack-dev-multi-stack", metav1.GetOptions{})
				return err == nil
			}, 300*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})

		ginkgo.By("kusion destroy", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "multi-stack", "dev")
			_, err := ExecKusionWithWorkDir("kusion destroy -y=true", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("wait multi-stack destroy", func() {
			homedir := os.Getenv("HOME")
			configPath := fmt.Sprintf("%s/.kube/config", homedir)
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.CoreV1().Namespaces().Get(context.TODO(), "multi-stack", metav1.GetOptions{})
				return apierrors.IsNotFound(err)
			}, 300*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})
	})
})

var _ = ginkgo.Describe("Kusion Other Commands", func() {
	ginkgo.Context("kusion env testing", func() {
		ginkgo.It("kusion env", func() {
			output, err := ExecKusion("kusion env")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("KUSION_PATH"))
		})

		ginkgo.It("kusion env json", func() {
			output, err := ExecKusion("kusion env --json")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("KUSION_PATH"))
		})
	})

	ginkgo.Context("kusion version testing", func() {
		ginkgo.It("kusion version", func() {
			output, err := ExecKusion("kusion version")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("releaseVersion"))
		})
	})
})
