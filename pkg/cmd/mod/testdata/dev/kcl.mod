[package]
name = "testdata"
version = "0.1.0"

[dependencies]
kam = { git = "https://github.com/KusionStack/kam.git", tag = "0.2.0" }
service = { oci = "oci://ghcr.io/kusionstack/service", tag = "0.1.0" }

[profile]
entries = ["../base/base.k", "main.k"]
