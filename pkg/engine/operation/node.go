package operation

type BaseNode struct {
	ID string
}

func (b *BaseNode) Hashcode() interface{} {
	return b.ID
}

func (b *BaseNode) Name() string {
	return b.ID
}

type RootNode struct{}

func (r *RootNode) Hashcode() interface{} {
	return "RootNode"
}

func (r *RootNode) Name() string {
	return "root"
}
