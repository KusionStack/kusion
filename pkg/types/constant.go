package types

type ResourceType string

const (
	K8SObject = ResourceType("K8SObject")
)

type ResourceAction string

const (
	ResourceGet      = ResourceAction("Get")
	ResourceCreate   = ResourceAction("Create")
	ResourceApply    = ResourceAction("Apply")
	ResourceDelete   = ResourceAction("Delete")
	ResourceAnnotate = ResourceAction("Annotate")
	ResourceDescribe = ResourceAction("Describe")
	ResourceLabel    = ResourceAction("Label")
	ResourcePatch    = ResourceAction("Patch")
	ResourceDiff     = ResourceAction("Diff")
)

const (
	EntranceFileName = "main.k"
)
