# Contributing to Kusion

Welcome to Kusion! Kusion is a cloud-native application delivery and operation engine, which is designed to simplify and optimize the process of delivering applications for dev and platform. Kusion belongs to [KusionStack](https://github.com/KusionStack), which is a technology stack in the field of platform engineering.

We warmly appreciate your talent and creativity in contributing to this project. This guide will help direct you to the best places to start contributing. Follow the instructions below, you'll be able to pick up issues, write code to fix them, and get your work reviewed and merged.

Feel free to create issues and contribute your code. Whether you are an experienced developer or just beginning your journey in the open-source world, we highly encourage your participation.

If you have any questions or need further information, please don't hesitate to contact us.

---

<!-- MarkdownTOC autolink="true" -->

- [File an Issue](#file-an-issue)
- [Contribute a Pull Request](#contribute-a-pull-request)
    - [Fork Repository](#fork-repository)
    - [Develop Code](#develop-code)
    - [Open a Pull Request](#open-a-pull-request)
    - [Sign CLA](#sign-cla)
    - [PR Checks](#pr-checks)

## File an Issue
The first step to start contributing to Kusion is to find something to work on. Help is always welcome, and no contribution is too small!

Here are some things you can do today to get started contributing:

* Help improve the Kusion documentation
* Clarify code, variables, or functions that can be renamed or commented on
* Write test coverage
* Help triage issues

We use [issues](https://github.com/KusionStack/kusion/issues) to track tasks. Choose an existed issue with label `good first issue` or`help wanted` is a good choice, or you can open a new issue. Now, Kusion provides three issue templates as follows, please choose one according to your need:

* Bug Report: Report a bug encountered while operating Kusion
* Enhancement Tracking Issue: Provide supporting details for a feature in development
* Failing Test: Report continuously failing tests or jobs in Kusion CI

In the issue, please describe your problem clearly and accurately, and choose a label to define the issue's type. There are the commonly used label:

* area/cli: Categorizes an issue or PR as relevant to CLI
* area/engine: Categorizes an issue or PR as relevant to Kusion engine
* area/release: Categorizes an issue or PR as relevant to build
* area/runtime: All Infrastructure Runtime, eg: k8s
* area/testing: Categorizes an issue or PR as relevant to test
* kind/feature: Categorizes issue or PR as related to a new feature

## Contribute a Pull Request

After opening an issue, you could contribute codes to Kusion by a pull request. Here are the steps you should follow:

### Fork Repository

Kusion adopts trunk-based development, i.e., the code used for release is maintained on the main branch. And to keep the repository tidy, Kusion only has the main branch. 

Thus, to develop Kusion, you have to fork the [KusionStack/kusion](https://github.com/KusionStack/kusion) repository to your own workspace, then checkout a new branch to develop coding.

### Develop Code

Now you can start coding to solve the issue. To maintain the code quality of Kusion, unit tests are indispensable for the functions you added or updated.

After the development is completed, commit and push to your own forked repository. Cause the code will merge into Kusion repository by pull request, [commitlint/config-conventional](https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional) should get followed, the legal prefixes for commit message are shown below:

* build
* chore
* ci 
* docs 
* feat 
* fix 
* perf 
* refactor 
* revert 
* style 
* test

### Open a Pull Request

[Open a pull request](https://github.com/KusionStack/kusion/pulls) from the develop branch of your forked repository to the main branch of Kusion. You should clearly describe what you do in the PR, and link it to an issue. Besides, the PR title should also follow the commit conventions described above, and must be 5-256 characters in length, prefix `WIP` and `[WIP]` are not allowed.

### Sign CLA

If it was your first pull request, you need to sign our [CLA(Contributor License Agreement)](https://github.com/KusionStack/.github/blob/main/CLA.md). The only thing you need to do is to post a pull request comment same as the below format:

`I have read the CLA Document and I hereby sign the CLA`

If your CLA signature failed, you may find the solutions as below:

* The comment must be in exactly the same format as above, with no extra spaces, line breaks, etc.
* The git committer must be the same one who create the Kusion PR

### PR Checks

To keep the reliability of the Kusion project, the following check will get triggered automatically:

* Unit Test
* E2E Test
* Golang Lint
* Commit Lint
* PR Title lint

Please make sure your PR passes these checks.