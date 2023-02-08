package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"kusionstack.io/kusion/third_party/kubevela/kubevela/apis/common"
)

// Package type metadata.
const (
	Group   = common.Group
	Version = "v1beta1"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}
