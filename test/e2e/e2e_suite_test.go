package e2e

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "E2e Suite")
}

// BeforeSuite Create kubernetes
var _ = ginkgo.BeforeSuite(func() {
	ginkgo.By("create k3s cluster", func() {
		cli := "k3d cluster create kusion-e2e"
		output, err := Exec(cli)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(output).To(gomega.ContainSubstring("successfully"))
	})

	ginkgo.By("git clone konfig", func() {
		output, err := ExecWithWorkDir("git clone https://github.com/KusionStack/konfig.git", GetWorkDir())
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(output).To(gomega.ContainSubstring("Cloning"))
	})

	ginkgo.By("kusion init code-city", func() {
		path := filepath.Join(GetWorkDir(), "konfig")
		output, err := ExecKusionWithStdin("kusion init --online=true --yes=true", path, "\n")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(output).To(gomega.ContainSubstring("Created project"))
	})
})

// AfterSuite clean kubernetes
var _ = ginkgo.AfterSuite(func() {
	ginkgo.By("clean up k3s cluster", func() {
		cli := "k3d cluster delete kusion-e2e"
		output, err := Exec(cli)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(output).To(gomega.ContainSubstring("Successfully"))
	})

	ginkgo.By("clean up konfig", func() {
		path := filepath.Join(GetWorkDir(), "konfig")
		cli := fmt.Sprintf("rm -rf %s", path)
		output, err := Exec(cli)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(output).To(gomega.BeEmpty())
	})

	ginkgo.By("clean up kusion e2e test binary", func() {
		path := filepath.Join(GetWorkDir(), "../..", "bin")
		cli := fmt.Sprintf("rm -rf %s", path)
		output, err := Exec(cli)
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
		gomega.Expect(output).To(gomega.BeEmpty())
	})
})
