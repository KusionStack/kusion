[package]
name = "{{ .ProjectName }}"
edition = "0.5.0"
version = "0.1.0"

[dependencies]
catalog = { git = "https://github.com/KusionStack/catalog.git", tag = "0.1.0" }

[profile]
entries = ["main.k"]
