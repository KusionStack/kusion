import catalog.models.schema.v1 as ac
import catalog.models.schema.v1.workload as wl
import catalog.models.schema.v1.workload.container as c
import catalog.models.schema.v1.workload.network as n

# main.k declares the customized configuration codes for dev stack. 
quickstart: ac.AppConfiguration {
    workload: wl.Service {
        containers: {
            quickstart: c.Container {
                image: "kusionstack/kusion-quickstart:latest"
            }
        }
        ports: [
            n.Port {
                port: 8080
            }
        ]
    }
}
