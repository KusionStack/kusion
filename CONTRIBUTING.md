# Contributing

Welcome to Kusion!

We warmly appreciate your talent and creativity in contributing to this project. This guide will help direct you to the best places to start contributing. Follow the instructions below, you'll be able to pick up issues, write code to fix them, and get your work reviewed and merged.

Feel free to create issues and contribute your code. Whether you are an experienced developer or just beginning your journey in the open-source world, we highly encourage your participation.

If you have any questions or need further information, please don't hesitate to contact us.

---

-   [Before you get started](#before-you-get-started)
    -   [Code of Conduct](#code-of-conduct)
    -   [Community Expectations](#community-expectations)
-   [Getting started](#getting-started)
-   [Your First Contribution](#your-first-contribution)
    -   [Find something to work on](#find-something-to-work-on)
        -   [Find a good first topic](#find-a-good-first-topic)
        -   [Work on an Issue](#work-on-an-issue)
-   [Contributor Workflow](#contributor-workflow)
    -   [Creating Pull Requests](#creating-pull-requests)
    -   [Code Review](#code-review)
    -   [Testing](#testing)

# Before you get started

## Code of Conduct

Please make sure to read our [Code of Conduct](./CODE_OF_CONDUCT.md).

## Community Expectations

Kusion is an open source project driven by its community which strives to promote a healthy, friendly and productive environment. Kusion aims to simplify the process of deploying applications into your infrastructure and helps you standardize your deployment process.

# Getting started

- Fork the repository on GitHub.
- Make your changes on your fork repository.
- Submit a PR.

# Your First Contribution

We will help you to contribute in different areas like filing issues, developing features, fixing critical bugs and getting your work reviewed and merged.

If you have questions about the development process, feel free to [file an issue](https://github.com/KusionStack/kusion/issues/new/choose).

## Find something to work on

We are always in need of help, be it fixing documentation, reporting bugs or writing some code.
Look at places where you feel best coding practices aren't followed, code refactoring is needed or tests are missing. Here is how you get started.

### Find a good first issue

[Kusion](https://github.com/KusionStack/kusion) has [help wanted](https://github.com/KusionStack/kusion/issues?q=is%3Aopen+is%3Aissue+label%3A%22help+wanted%22) and [good first issue](https://github.com/KusionStack/kusion/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22) labels for issues that should not need deep knowledge of the system. We can help new contributors who wish to work on such issues.

Another good way to contribute is to find a documentation improvement, such as a missing/broken link.
Please see [Contributing](#contributing) below for the workflow.

#### Work on an issue

When you are willing to take on an issue, just reply on the issue. The maintainer will assign it to you.

# Contributor Workflow

Please do not ever hesitate to ask a question or send a pull request.

This is a rough outline of what a contributor's workflow looks like:

- Create a topic branch from where to base the contribution. This is usually master.
- Make commits of logical units.
- Push changes in a topic branch to a personal fork of the repository.
- Submit a pull request to [Kusion](https://github.com/KusionStack/kusion).

## Creating Pull Requests

Pull requests are often called simply "PR". Kusion generally follows the standard [github pull request](https://help.github.com/articles/about-pull-requests/) process. To submit a proposed change, please develop the code/fix and add new test cases. After that, run these local verifications before submitting pull request to predict the pass or fail of continuous integration.

* Run and pass `make lint`
* Run and pass `make test`

## Code Review

To make it easier for your PR to receive reviews, consider the reviewers will need you to:

* follow [good coding guidelines](https://github.com/golang/go/wiki/CodeReviewComments).
* write [good commit messages](https://chris.beams.io/posts/git-commit/).
* break large changes into a logical series of smaller patches which individually make easily understandable changes, and in aggregate solve a broader issue.
