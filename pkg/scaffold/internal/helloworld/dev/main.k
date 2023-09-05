import catalog.models.schema.v1 as ac

# main.k declares customized configurations for dev stack.
{{ .AppName }}: ac.AppConfiguration {
    workload.containers.{{ .AppName }}: {
        # dev stack has different image
        image = "{{ .Image }}"
    }
}
