package tfops

import (
	"context"
	"fmt"
	"testing"
)

func TestCreate(t *testing.T) {
	type args struct {
		ws *WorkspaceStore
	}

	tests := map[string]struct {
		args
	}{
		"Success": {
			args: args{
				ws: &WorkspaceStore{
					Store: make(map[string]*WorkSpace),
					Fs:    fs,
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tt.args.ws.Create(context.TODO(), &resourceTest); err != nil {
				t.Errorf("\n workspaceStore Create error: %v", err)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	type args struct {
		ws *WorkspaceStore
	}

	tests := map[string]struct {
		args
	}{
		"SuccessRemove": {
			args: args{
				ws: &WorkspaceStore{
					Store: make(map[string]*WorkSpace),
					Fs:    fs,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tt.args.ws.Remove(context.TODO(), &resourceTest); err != nil {
				t.Errorf("\n workspaceStore Remove error: %v", err)
			}
		})
	}
}

func TestGetWorkspaceStore(t *testing.T) {
	type args struct {
		ws *WorkspaceStore
	}

	tests := map[string]struct {
		args
	}{
		"GetworkspaceStore": {
			args: args{
				&WorkspaceStore{
					Store: make(map[string]*WorkSpace),
					Fs:    fs,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ws, err := GetWorkspaceStore(tt.ws.Fs)
			fmt.Println(ws)
			if err != nil {
				t.Errorf("\nGetWorkspaceStore error: %v", err)
			}
		})
	}
}
