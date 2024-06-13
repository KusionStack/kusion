package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = ginkgo.Describe("Kusion Configuration Commands", func() {
	ginkgo.Context("kusion generate testing", func() {
		ginkgo.It("kusion generate", func() {
			// kusion build testing
			path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
			output, err := ExecKusionWithWorkDir("kusion generate", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("Generating Spec"))
		})
	})
})

var _ = ginkgo.Describe("kusion Runtime Commands", func() {
	ginkgo.It("kusion preview", func() {
		path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
		_, err := ExecKusionWithWorkDir("kusion preview -d=false", path)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})

	ginkgo.It("kusion apply", func() {
		ginkgo.By("kusion apply", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
			_, err := ExecKusionWithWorkDir("kusion apply -y=true", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("wait service-multi-stack deploy", func() {
			homedir := os.Getenv("HOME")
			configPath := fmt.Sprintf("%s/.kube/config", homedir)
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.AppsV1().Deployments("service-multi-stack").Get(context.TODO(), "service-multi-stack-dev-echoserver", metav1.GetOptions{})
				return err == nil
			}, 300*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})

		ginkgo.By("kusion destroy", func() {
			path := filepath.Join(GetWorkDir(), "konfig", "example", "service-multi-stack", "dev")
			_, err := ExecKusionWithWorkDir("kusion destroy -y=true", path)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		})

		ginkgo.By("wait service-multi-stack destroy", func() {
			homedir := os.Getenv("HOME")
			configPath := fmt.Sprintf("%s/.kube/config", homedir)
			clusterConfig, err := clientcmd.BuildConfigFromFlags("", configPath)
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			clusterClient := kubernetes.NewForConfigOrDie(clusterConfig)
			gomega.Eventually(func() bool {
				_, err := clusterClient.CoreV1().Namespaces().Get(context.TODO(), "service-multi-stack", metav1.GetOptions{})
				return errors.IsNotFound(err)
			}, 300*time.Second, 5*time.Second).Should(gomega.Equal(true))
		})
	})
})

var _ = ginkgo.Describe("Kusion Other Commands", func() {
	ginkgo.Context("kusion version testing", func() {
		ginkgo.It("kusion version", func() {
			output, err := ExecKusion("kusion version")
			gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
			gomega.Expect(output).To(gomega.ContainSubstring("releaseVersion"))
		})
	})
})
