package compile

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestGetKclPath(t *testing.T) {
	os.Setenv(KusionKclPathEnv, "kcl-custom-path")
	tAssert(t, getKclPath() == "kcl-custom-path")

	os.Setenv(KusionKclPathEnv, "")

	_ = os.MkdirAll("./kclvm/bin", 0o777)

	kclData := fmt.Sprintf("# kcl-test shell, %d", time.Now().Unix())
	_ = os.WriteFile("./kclvm/bin/kcl", []byte(kclData), 0o777)
	defer os.RemoveAll("./kclvm")

	kcl := getKclPath()
	kclDataGot, _ := os.ReadFile(kcl)
	if len(kclDataGot) > 50 {
		kclDataGot = kclDataGot[:50]
	}
	tAssert(t, kclData == string(kclDataGot), kclData, string(kclDataGot))
	os.RemoveAll("./kclvm")

	if s, _ := exec.LookPath("kcl"); s != "" {
		kcl := getKclPath()
		tAssert(t, kcl == s, s, kcl)
	}
}

func tAssert(tb testing.TB, condition bool, a ...interface{}) {
	tb.Helper()
	if !condition {
		if msg := fmt.Sprint(a...); msg != "" {
			tb.Fatal("Assert failed: " + msg)
		} else {
			tb.Fatal("Assert failed")
		}
	}
}
