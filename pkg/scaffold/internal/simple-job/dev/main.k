import catalog.models.schema.v1 as ac
import catalog.models.schema.v1.workload as wl
import catalog.models.schema.v1.workload.container as c

{{ .AppName }}: ac.AppConfiguration {
    workload: wl.Job {
        containers: {
            "{{ .AppName }}": c.Container {
                image: "{{ .Image }}"
                # The job will run the following command.
                command: ["/bin/sh", "-c", "echo hello"]
            }
        }
        # The job will run every hour.
        schedule: "0 * * * *"
    }
}
