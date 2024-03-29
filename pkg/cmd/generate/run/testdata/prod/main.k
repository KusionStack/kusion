import kam.v1.app_configuration as ac

# main.k declares customized configurations for prod stack.
helloworld: ac.AppConfiguration {
    workload.containers.nginx: {
        # prod stack has different image
        image = "nginx:v2"
    }
}