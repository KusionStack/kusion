import kam.v1.app_configuration as ac
import kam.v1.workload as wl
import kam.v1.workload.container as c
import network as n

# main.k declares the customized configuration codes for dev stack. 
quickstart: ac.AppConfiguration {
    workload: wl.Service {
        containers: {
            quickstart: c.Container {
                image: "kusionstack/kusion-quickstart:latest"
            }
        }
    }
    accessories: {
        "network": n.Network {
            ports: [
                n.Port {
                    port: 8080
                }
            ]
        }
    }
}
