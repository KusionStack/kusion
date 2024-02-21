[package]
name = "testdata"
version = "0.1.0"

[dependencies]
catalog = { git = "https://github.com/KusionStack/catalog.git", tag = "0.1.2" }
[profile]
entries = ["../base/base.k", "main.k"]

