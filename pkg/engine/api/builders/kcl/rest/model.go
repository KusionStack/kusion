package rest

import (
	"kcl-lang.io/kcl-go/pkg/spec/gpyrpc"
)

type Result struct {
	Error  string                     `json:"error"`
	Result *gpyrpc.ExecProgram_Result `json:"result"`
}

type PingResponse struct {
	Error  string      `json:"error"`
	Result *PingResult `json:"result"`
}

type PingResult struct {
	Value string `json:"value,omitempty"`
}
