# `/third_party`

## `/diff`

This package provides functions that allow to compare set of Kubernetes resources using the logic equivalent to `kubectl diff`.

Source code is developed secondary based on GitHub repo [https://github.com/argoproj/gitops-engine](https://github.com/argoproj/gitops-engine),
version `v0.5.2`, you may check them under package `pkg/diff`.

A few changes made by KusionStack:
- `options.go` is part of `diff_options.go`.
- `diff_normalizer.go` is newly developed, which provides a `ignoreNormalizer` to ignore fields according to given json path.

## `/dyff`

Similar to the standard `diff` tool, it follows the principle of describing the change by going from the `from` input file to the target `to` input file.

Source code mainly comes from GitHub repo [https://github.com/homeport/dyff](https://github.com/homeport/dyff),
version `v1.1.0`, you may check them under package `pkg/dyff`.

A few changes made by KusionStack:
- `custom_comparator.go` provide a map of special fields and its comparator function, which is injected into report of `CompareInputFiles`.

## `/pulumi`

- `fsutil` provides a util function to walk up each file in specified path.
- `gitutil` and `workspace` provides some util functions to simplify git related operations.

Source code mainly comes from GitHub repo [https://github.com/pulumi/pulumi](https://github.com/pulumi/pulumi),
version `v3.24.0` you may check them under package `sdk/go/common`.
