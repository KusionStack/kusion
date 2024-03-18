FROM golang:1.22 AS build
COPY / /src
WORKDIR /src
RUN --mount=type=cache,target=/go/pkg --mount=type=cache,target=/root/.cache/go-build make build-local-linux

FROM ubuntu:22.04 AS base
# Install KCL Dependencies
RUN apt-get update -y && apt-get install python3 python3-pip -y
# KCL PATH
ENV PATH="/root/go/bin:${PATH}"
# KUSION
ENV KUSION_HOME="$HOME/.kusion"
ENV LANG=en_US.utf8

FROM base AS goreleaser
COPY kusion /usr/local/bin/kusion
RUN /usr/local/bin/kusion

FROM base
COPY --from=build /src/_build/bundles/kusion-linux/bin/kusion /usr/local/bin/kusion
RUN /usr/local/bin/kusion