package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// 默认下载当前系统对应的最新版本
// https://pypi.antfin-inc.com/
const (
	kclLinuxUrl_latest       = "https://pypi.antfin-inc.com/api/package/kclvm-dev/kclvm-dev-linux-latest.tar.gz"
	kclDarwinUrl_latest      = "https://pypi.antfin-inc.com/api/package/kclvm-dev/kclvm-dev-Darwin-latest.tar.gz"
	kclDarwinArm64Url_latest = "https://pypi.antfin-inc.com/api/package/kclvm-dev/kclvm-dev-Darwin-arm64-latest.tar.gz"
	kclWindowsUrl_latest     = "https://pypi.antfin-inc.com/api/package/kclvm-dev/kclvm-dev-windows-latest.tar.gz"
	kclAlpineUrl_latest      = "https://pypi.antfin-inc.com/api/package/kclvm-dev/kclvm-dev-alpine-latest.tar.gz"
)

var (
	flagKclUrl       = flag.String("kcl-url", "", "set kcl url")
	flagInstallPath  = flag.String("install-path", ".", "set install path")
	flagKclPlatform  = flag.String("kcl-platform", "", "set kcl platform")
	flagIsKCLOpenAPI = flag.Bool("is-kclopenapi", false, "is download KCLOpenAPI")
)

func main() {
	flag.Parse()
	if *flagKclUrl == "" {
		if *flagKclPlatform == "" {
			*flagKclPlatform = runtime.GOOS
		}
		switch *flagKclPlatform {
		case "linux":
			*flagKclUrl = kclLinuxUrl_latest
		case "darwin":
			*flagKclUrl = kclDarwinUrl_latest
		case "darwin-arm64":
			*flagKclUrl = kclDarwinArm64Url_latest
		case "windows":
			*flagKclUrl = kclWindowsUrl_latest
		case "alpine":
			*flagKclUrl = kclAlpineUrl_latest
		default:
			log.Fatal("empty url")
		}
	}
	if *flagInstallPath == "" {
		*flagInstallPath = "."
	}

	// 下载 tar.gz 文件
	data, err := httpGet(*flagKclUrl)
	if err != nil {
		log.Fatal(err)
	}

	if *flagIsKCLOpenAPI {
		// Copy KCLOpenAPI to install path
		err = installKCLOpenAPI(data, *flagInstallPath)
	} else {
		// 展开到本地目录
		err = installKclvm(data, *flagInstallPath)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func installKCLOpenAPI(kclOpenAPI []byte, installPath string) error {
	reader := bytes.NewReader(kclOpenAPI)
	outFile, err := os.Create(filepath.Join(installPath, "kclopenapi"))
	if err != nil {
		return err
	}
	if _, err := io.Copy(outFile, reader); err != nil {
		return err
	}
	outFile.Close()
	return nil
}

func installKclvm(kclTarGz []byte, installPath string) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(kclTarGz))
	if err != nil {
		return err
	}

	// 需要保留的文件
	goodPath := func(path string) string {
		// kclvm-dev-0.2.0/kclvm/bin/...
		if idx := strings.Index(path, "kclvm/"); idx >= 0 {
			return installPath + "/" + path[idx:]
		}
		if strings.HasSuffix(path, "/README.KCL.md") {
			return installPath + "/kclvm/" + "README.KCL.md"
		}
		if strings.HasSuffix(path, "/VERSION") {
			return installPath + "/kclvm/" + "VERSION"
		}
		return ""
	}

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if s := goodPath(header.Name); s != "" {
				os.MkdirAll(header.Name, 0o755)
			}
		case tar.TypeReg:
			if s := goodPath(header.Name); s != "" {
				os.MkdirAll(filepath.Dir(s), 0o755)
				outFile, err := os.Create(s)
				if err != nil {
					return err
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					return err
				}
				outFile.Close()
			} else {
				io.Copy(ioutil.Discard, tarReader)
			}
		}
	}
}

func httpGet(url string) (data []byte, err error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download %s: %v", url, err)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func fileExists(filename string) bool {
	fi, err := os.Stat(filename)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}
