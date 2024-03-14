package e2e

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	v1 "kusionstack.io/kusion/pkg/apis/core/v1"
	"kusionstack.io/kusion/pkg/workspace"
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

	ginkgo.By("kusion init", func() {
		path := filepath.Join(GetWorkDir(), "konfig")
		_, err := ExecKusionWithStdin("kusion init --online=true --yes=true", path, "\n")
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})

	ginkgo.By("create sample workspace", func() {
		err := createSampleWorkspace()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
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

	ginkgo.By("clean up sample workspace", func() {
		err := deleteSampleWorkspace()
		gomega.Expect(err).ShouldNot(gomega.HaveOccurred())
	})
})

// createSampleWorkspace creates a sample workspace under default local path.
// todo: 1, add REAL sample workspace in repo Konfig after splitting work; 2, use cli to create workspace after cli R&D work.
func createSampleWorkspace() error {
	wsOperator, err := workspace.NewDefaultOperator()
	if err != nil {
		return err
	}
	ws := &v1.Workspace{
		Name: "dev",
		// comment out for now, as download modules from remote repo is not supported yet.
		// Modules: v1.ModuleConfigs{
		//	"database": {
		//		Default: v1.GenericConfig{
		//			"type":         "aws",
		//			"version":      "5.7",
		//			"instanceType": "db.t3.micro",
		//		},
		//		ModulePatcherConfigs: v1.ModulePatcherConfigs{
		//			"smallClass": {
		//				GenericConfig: v1.GenericConfig{
		//					"instanceType": "db.t3.small",
		//				},
		//				ProjectSelector: []string{"foo", "bar"},
		//			},
		//		},
		//	},
		//	"port": {
		//		Default: v1.GenericConfig{
		//			"type": "aws",
		//		},
		//	},
		// },
	}
	return wsOperator.CreateWorkspace(ws)
}

// deleteSampleWorkspace deletes the sample workspace.
func deleteSampleWorkspace() error {
	wsOperator, err := workspace.NewDefaultOperator()
	if err != nil {
		return err
	}
	return wsOperator.DeleteWorkspace("dev")
}
