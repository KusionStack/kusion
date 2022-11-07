FROM ubuntu:20.04

COPY _build/bundles/kusion-linux/ /kusion/
COPY _build/bundles/kusion-ubuntu/ /kusion/

RUN chmod +x /kusion/bin/kusion \
&&  chmod +x /kusion/bin/kcl-openapi \
&&  chmod +x /kusion/kclvm/bin/kcl \
&&  chmod +x /kusion/kclvm/bin/kclvm \
&&  chmod +x /kusion/kclvm/bin/kcl-doc \
&&  chmod +x /kusion/kclvm/bin/kcl-plugin \
&&  chmod +x /kusion/kclvm/bin/kcl-test \
&&  chmod +x /kusion/kclvm/bin/kcl-lint \
&&  chmod +x /kusion/kclvm/bin/kcl-fmt \
&&  chmod +x /kusion/kclvm/bin/kcl-vet \
&&  chmod +x /kusion/kclvm/bin/kcl-go \
&&  chmod +x /kusion/kclvm/bin/kclvm_cli

# Install dependency
RUN apt-get update -y
RUN apt-get install -y clang-12 lld-12 libssl-dev --no-install-recommends
RUN apt-get clean all
RUN ln -sf /usr/bin/clang-12   /usr/bin/clang
RUN ln -sf /usr/bin/wasm-ld-12 /usr/bin/wasm-ld

ENV PATH="/kusion/bin:/kusion/kclvm/bin:${PATH}"
ENV KUSION_PATH="/kusion"
ENV LANG=en_US.utf8
