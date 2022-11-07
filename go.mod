module kusionstack.io/kusion

go 1.16

require (
	bou.ke/monkey v1.0.2
	github.com/AlecAivazis/survey/v2 v2.3.4
	github.com/Azure/go-autorest/autorest/mocks v0.4.1
	github.com/aliyun/aliyun-oss-go-sdk v2.1.8+incompatible
	github.com/aws/aws-sdk-go v1.42.35
	github.com/blang/semver/v4 v4.0.0
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1
	github.com/davecgh/go-spew v1.1.1
	github.com/didi/gendry v1.7.0
	github.com/djherbis/times v1.5.0
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484 // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-errors/errors v1.4.0 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/goccy/go-yaml v1.8.9
	github.com/gonvenience/bunt v1.1.1
	github.com/gonvenience/neat v1.3.0
	github.com/gonvenience/term v1.0.0
	github.com/gonvenience/text v1.0.5
	github.com/gonvenience/wrap v1.1.0
	github.com/gonvenience/ytbx v1.3.0
	github.com/google/go-cmp v0.5.8
	github.com/gookit/goutil v0.5.1
	github.com/gosuri/uilive v0.0.4
	github.com/hashicorp/go-version v1.4.0
	github.com/hashicorp/hcl/v2 v2.11.1
	github.com/hashicorp/terraform v0.15.3
	github.com/imdario/mergo v0.3.13
	github.com/jinzhu/copier v0.3.2
	github.com/lucasb-eyer/go-colorful v1.0.3
	github.com/mitchellh/hashstructure v1.0.0
	github.com/onsi/ginkgo/v2 v2.0.0
	github.com/onsi/gomega v1.18.1
	github.com/pkg/errors v0.9.1
	github.com/pterm/pterm v0.12.42-0.20220427210824-6bb8c6e6cc77
	github.com/sergi/go-diff v1.2.0
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.1
	github.com/texttheater/golang-levenshtein v1.0.1
	github.com/zclconf/go-cty v1.10.0
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa // indirect
	golang.org/x/sys v0.0.0-20220808155132-1c4a2a72c664 // indirect
	golang.org/x/term v0.0.0-20220411215600-e5f449aeb171 // indirect
	google.golang.org/genproto v0.0.0-20210420162539-3c870d7478d2 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0
	k8s.io/api v0.24.2
	k8s.io/apimachinery v0.24.2
	k8s.io/client-go v10.0.0+incompatible
	k8s.io/component-base v0.24.2
	k8s.io/kubectl v0.21.2
	kusionstack.io/kcl-plugin v0.4.1-alpha2
	kusionstack.io/kclvm-go v0.4.3
	sigs.k8s.io/controller-runtime v0.12.3
	sigs.k8s.io/kustomize/api v0.8.11
	sigs.k8s.io/kustomize/kustomize/v4 v4.1.2
	sigs.k8s.io/kustomize/kyaml v0.11.0
	sigs.k8s.io/yaml v1.3.0
)

replace (
	github.com/gosuri/uilive => github.com/howieyuen/uilive v0.0.5
	k8s.io/api => k8s.io/api v0.21.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.2
	k8s.io/apiserver => k8s.io/apiserver v0.21.2
	k8s.io/client-go => k8s.io/client-go v0.21.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.2
	k8s.io/component-base => k8s.io/component-base v0.21.2
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.2
	k8s.io/cri-api => k8s.io/cri-api v0.21.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.2
	k8s.io/kubelet => k8s.io/kubelet v0.21.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.2
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.2
	sigs.k8s.io/kustomize/kustomize/v4 => sigs.k8s.io/kustomize/kustomize/v4 v4.2.0
)
