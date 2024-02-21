import catalog.models.schema.v1 as ac

# main.k declares customized configurations for prod stack.
helloworld: ac.AppConfiguration {
    workload.containers.nginx: {
        # prod stack has different image
        image = "nginx:v2"
    }
}

sampleapp: ac.AppConfiguration {
    workload.containers.nginx: {
        # prod stack has different image
        image = "nginx:v3"
    }
}