FROM ubuntu:20.04

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

# Install KCL Dependencies
RUN apt-get update -y && apt-get install python3 python3-pip -y
RUN /kusion/kclvm/bin/kcl

ENV PATH="/kusion/bin:/kusion/kclvm/bin:${PATH}"
ENV KUSION_PATH="/kusion"
ENV LANG=en_US.utf8
