import catalog.models.schema.v1 as ac
import catalog.models.schema.v1.trait as t
import catalog.models.schema.v1.workload as wl
import catalog.models.schema.v1.workload.container as c
import catalog.models.schema.v1.workload.container.probe as p
import catalog.models.schema.v1.workload.secret as sec
import catalog.models.schema.v1.workload.network as n

# base.k declares reuseable configurations for all stacks.
{{ .AppName }}: ac.AppConfiguration {
    workload: wl.Service {
        containers: {
            "{{ .AppName }}": c.Container {
                resources: {
                    "cpu": "2"
                    "memory": "4Gi"
                }
            }
        }
        replicas: 2
        ports: [
            n.Port {
                port: 80
            }
        ]
    }
}

