# `/scripts`

+ *install.sh*: install kusion components, kclvm and kusionCtl binary.

+ *nvm_install.sh*: used by *Dockerfile_CloudIDE_Platform_node14*, install nvm to $HOME/.nvm
+ *Dockerfile_CloudIDE_Platform_node14*: the Dockerfile of the cloudide base imgae with node14 installed. It add node14 layer to the official cloudide platform base image.

+ *set_aci_env.sh*: used by antcode ci pipeline. It init the build environment of the aci job.