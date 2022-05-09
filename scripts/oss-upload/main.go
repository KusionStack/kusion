//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"

	"kusionstack.io/kusion/pkg/client/oss"
)

const (
	ENV_OSS_ACCESS_KEY_ID                  = "OSS_ACCESS_KEY_ID"
	ENV_OSS_ACCESS_KEY_SECRET              = "OSS_ACCESS_KEY_SECRET"
	ENV_OSS_ENDPOINT                       = "OSS_ENDPOINT"
	ENV_OSS_BUCKET_NAME                    = "OSS_BUCKET_NAME"
	ENV_OSS_KUSION_DARWIN_BUCKET_KEY       = "OSS_KUSION_DARWIN_BUCKET_KEY"
	ENV_OSS_KUSION_DARWIN_ARM64_BUCKET_KEY = "OSS_KUSION_DARWIN_ARM64_BUCKET_KEY"
	ENV_OSS_KUSION_LINUX_BUCKET_KEY        = "OSS_KUSION_LINUX_BUCKET_KEY"
	ENV_OSS_KUSION_WINDOWS_BUCKET_KEY      = "OSS_KUSION_WINDOWS_BUCKET_KEY"

	kusionDarwinReleasePath      = "build/bundles/kusion-darwin.tgz"
	kusionDarwinArm64ReleasePath = "build/bundles/kusion-darwin-arm64.tgz"
	kusionLinuxReleasePath       = "build/bundles/kusion-linux.tgz"
	kusionWindowsReleasePath     = "build/bundles/kusion-windows.zip"
)

func main() {
	var id, secret, endPoint, bucketName string
	var kusionDarwinBucketKey, kusionDarwinArm64BucketKey, kusionLinuxBucketKey, kusionWindowsBucketKey string

	// Complete the values from environment variables.
	id = os.Getenv(ENV_OSS_ACCESS_KEY_ID)
	secret = os.Getenv(ENV_OSS_ACCESS_KEY_SECRET)

	if env := os.Getenv(ENV_OSS_ENDPOINT); env != "" {
		endPoint = env
	}

	if env := os.Getenv(ENV_OSS_BUCKET_NAME); env != "" {
		bucketName = env
	}

	if env := os.Getenv(ENV_OSS_KUSION_DARWIN_BUCKET_KEY); env != "" {
		kusionDarwinBucketKey = env
	}

	if env := os.Getenv(ENV_OSS_KUSION_DARWIN_ARM64_BUCKET_KEY); env != "" {
		kusionDarwinArm64BucketKey = env
	}

	if env := os.Getenv(ENV_OSS_KUSION_LINUX_BUCKET_KEY); env != "" {
		kusionLinuxBucketKey = env
	}

	if env := os.Getenv(ENV_OSS_KUSION_WINDOWS_BUCKET_KEY); env != "" {
		kusionWindowsBucketKey = env
	}

	client, err := oss.NewRemoteClient(endPoint, id, secret, "", bucketName)
	if err != nil {
		panic(err)
	}
	if fileExists(kusionDarwinReleasePath) {
		err = client.PutObjectFromFile(kusionDarwinBucketKey, kusionDarwinReleasePath)
		if err != nil {
			panic(err)
		}
		if fileExists(kusionDarwinReleasePath + ".md5.txt") {
			err = client.PutObjectFromFile(kusionDarwinBucketKey+".md5.txt", kusionDarwinReleasePath+".md5.txt")
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("Upload Kusion-darwin to OSS Successfully!")
	}
	if fileExists(kusionDarwinArm64ReleasePath) {
		err = client.PutObjectFromFile(kusionDarwinArm64BucketKey, kusionDarwinArm64ReleasePath)
		if err != nil {
			panic(err)
		}
		if fileExists(kusionDarwinArm64ReleasePath + ".md5.txt") {
			err = client.PutObjectFromFile(kusionDarwinArm64BucketKey+".md5.txt", kusionDarwinArm64ReleasePath+".md5.txt")
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("Upload Kusion-darwin-arm64 to OSS Successfully!")
	}
	if fileExists(kusionLinuxReleasePath) {
		err = client.PutObjectFromFile(kusionLinuxBucketKey, kusionLinuxReleasePath)
		if err != nil {
			panic(err)
		}
		if fileExists(kusionLinuxReleasePath + ".md5.txt") {
			err = client.PutObjectFromFile(kusionLinuxBucketKey+".md5.txt", kusionLinuxReleasePath+".md5.txt")
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("Upload Kusion-linux to OSS Successfully!")
	}
	if fileExists(kusionWindowsReleasePath) {
		err = client.PutObjectFromFile(kusionWindowsBucketKey, kusionWindowsReleasePath)
		if err != nil {
			panic(err)
		}
		if fileExists(kusionWindowsReleasePath + ".md5.txt") {
			err = client.PutObjectFromFile(kusionWindowsBucketKey+".md5.txt", kusionWindowsReleasePath+".md5.txt")
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("Upload Kusion-windows to OSS Successfully!")
	}
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
