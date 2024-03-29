[package]
name = "testdata"
version = "0.1.0"

[dependencies]
opsrule = { oci = "oci://ghcr.io/kusionstack/opsrule", tag = "0.1.0" }
kam = { git = "https://github.com/KusionStack/kam.git", tag = "0.1.0" }

[profile]
entries = ["../base/base.k", "main.k"]

