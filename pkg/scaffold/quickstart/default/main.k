import kam.v1.app_configuration as ac
import service
import service.container as c
import network as n

# main.k declares the customized configuration codes for default stack. 
quickstart: ac.AppConfiguration {
    workload: service.Service {
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
