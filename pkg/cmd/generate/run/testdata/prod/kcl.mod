[package]
name = "testdata"
version = "0.1.0"

[dependencies]
catalog = { path = "/Users/linkedin/Documents/go/src/github.com/kusionstack.io/catalog" }
opsrule = { oci = "oci://ghcr.io/kusionstack/opsrule", tag = "0.0.9" }

[profile]
entries = ["../base/base.k", "main.k"]

