// Copyright 2021 The Kusion Authors. All rights reserved.

//go:build ignore
// +build ignore

// 安装 go.mod 依赖的 kclvm 和 plugins
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	kclvm_py "kusionstack.io/KCLVM/kclvm"
	plugins "kusionstack.io/kcl_plugins"
)

var (
	flagKclvmRoot  = flag.String("kclvm-root", "", "set kclvm root")
	flagTargetOS   = flag.String("target-os", "", "set target os")
	flagTargetArch = flag.String("target-arch", "", "set target arch")
)

func main() {
	flag.Parse()

	if *flagKclvmRoot == "" {
		fmt.Println("ERR: -kclvm-root invalid")
		os.Exit(1)
	}

	switch *flagTargetOS {
	case "darwin":
		if s := *flagTargetArch; s != "amd64" && s != "arm64" {
			fmt.Printf("ERR: -target-arch invalid: %q\n", s)
			os.Exit(1)
		}

	case "linux":
		if s := *flagTargetArch; s != "amd64" {
			fmt.Printf("ERR: -target-arch invalid: %q\n", s)
			os.Exit(1)
		}
	case "windows":
		if s := *flagTargetArch; s != "amd64" {
			fmt.Printf("ERR: -target-arch invalid: %q\n", s)
			os.Exit(1)
		}
	default:
		fmt.Printf("ERR: -target-os invalid: %q\n", *flagTargetOS)
		os.Exit(1)
	}

	kclvmLibPath := getKclvmLibPath(*flagKclvmRoot)
	kclvmPluginsPath := getPluginsPath(*flagKclvmRoot)

	os.RemoveAll(kclvmLibPath)
	os.RemoveAll(kclvmPluginsPath)

	if err := kclvm_py.InstallKclvmLibs(kclvmLibPath); err != nil {
		fmt.Println("failed")
		fmt.Println("InstallKclvmLibs failed:", err)
		os.Exit(1)
	}
	if err := plugins.InstallPlugins(kclvmPluginsPath); err != nil {
		fmt.Println("InstallPlugins failed:", err)
		os.Exit(1)
	}

	fmt.Println("install kclvm lib & plugins ok")
}

// 获取 kclvm 包路径
func getKclvmLibPath(kclvmRoot string) string {
	switch *flagTargetOS {
	case "darwin", "linux":
		for _, python3_x := range []string{
			"python3.7",
			"python3.8",
			"python3.9",
			"python3.10",
		} {
			kclvmLibPath := filepath.Join(kclvmRoot, "lib", python3_x)
			if fi, _ := os.Stat(kclvmLibPath); fi != nil && fi.IsDir() {
				return filepath.Join(kclvmLibPath, "kclvm")
			}
		}
		panic(fmt.Sprintf("find lib failed: kclvmRoot = %s", kclvmRoot))

	case "windows":
		kclvmLibPath := filepath.Join(kclvmRoot, "bin", "Lib")
		return filepath.Join(kclvmLibPath, "kclvm")
	}

	panic("unreachable")
}

// 获取 plugins 路径
func getPluginsPath(kclvmRoot string) string {
	if *flagTargetOS == "windows" {
		kclvmPluginPath := filepath.Join(kclvmRoot, "bin", "plugins")
		return kclvmPluginPath
	}
	kclvmPluginPath := filepath.Join(kclvmRoot, "plugins")
	return kclvmPluginPath
}
