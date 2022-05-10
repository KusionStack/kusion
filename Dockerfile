FROM centos:centos7

COPY build/bundles/kusion-linux/ /kusion/

RUN chmod +x /kusion/bin/kusion \
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
RUN yum install -y wget git gcc

# KCLVMx Install dependency
# RUN yum -y install centos-release-scl
# RUN yum-config-manager --enable rhel-server-rhscl-7-rpms
# RUN yum -y install llvm-toolset-7.0
# RUN scl enable llvm-toolset-7.0 bash
# ENV LD_LIBRARY_PATH="/opt/rh/llvm-toolset-7.0/root/usr/lib64:${LD_LIBRARY_PATH}"
# ENV PATH="/opt/rh/llvm-toolset-7.0/root/usr/bin:${PATH}"

# Install ossutil
RUN wget -q -P /usr/local/bin http://gosspublic.alicdn.com/ossutil/1.7.5/ossutil64 \
&& chmod 755 /usr/local/bin/ossutil64

ENV PATH="/kusion/bin:/kusion/kclvm/bin:${PATH}"
ENV LANG=en_US.utf8
