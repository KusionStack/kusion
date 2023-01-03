FROM ubuntu:20.04

COPY _build/bundles/kusion-linux/ /kusion/

RUN chmod +x /kusion/bin/kusion

# Install KCL Dependencies
RUN apt-get update -y && apt-get install python3 python3-pip -y
# unembed kcl stuff
RUN /kusion/bin/kusion

ENV PATH="/kusion/bin:/root/go/bin:${PATH}"
ENV KUSION_PATH="/kusion"
ENV LANG=en_US.utf8
