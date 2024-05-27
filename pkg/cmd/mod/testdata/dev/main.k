import kam.v1.app_configuration as ac
import service as svc
import service.container as c

# base.k declares reusable configurations for all stacks.
helloworld: ac.AppConfiguration {
    workload: svc.Service {
        containers: {
            "nginx": c.Container {
                image: "nginx:v1"
                # Run the following command as defined
                command: ["/bin/sh", "-c", "echo hi"]
                # Extra arguments append to command defined above
                args: ["/bin/sh", "-c", "echo hi"]
                env: {
                    # An environment variable of name "env1" and value "VALUE" will be set
                    "env1": "VALUE"
                    # An environment variable of name "env2" and value of the key "key" in the
                    # secret named "sec-name" will be set.
                    "env2": "secret://sec-name/key"
                }
                # Run the command "/bin/sh -c echo hi", as defined above, in the directory "/tmp"
                workingDir: "/tmp"
            }
        }
        # Set the replicas
        replicas: 2
    }
}