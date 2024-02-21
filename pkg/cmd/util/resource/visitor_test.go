package resource

import (
	"testing"
)

func TestCrdVisitor(t *testing.T) {
	visitor := CrdVisitor{
		Path: "./crd",
	}
	objs, err := visitor.Visit()
	if err != nil {
		t.Errorf("unexpected err: %v", err)
	}
	if len(objs) != 2 {
		t.Errorf("unexpected doc size: %d", len(objs))
	}
}
