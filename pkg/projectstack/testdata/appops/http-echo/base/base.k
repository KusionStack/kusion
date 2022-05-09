import base.pkg.kusion_models.kube.frontend
import base.pkg.kusion_models.kube.frontend.container
import base.pkg.kusion_models.kube.templates.resource as res_tpl
import base.pkg.kusion_models.kube.frontend.service
import base.pkg.kusion_models.kube.frontend.ingress

# Application Configuration
appConfiguration: frontend.Server {
    # Main Container Configuration
    mainContainer = container.Main {
        name = "apple-app"
        args = ["-text=apple"]
    }
    image = "hashicorp/http-echo"
    schedulingStrategy.resource = res_tpl.medium
    services = [
        service.Service {
            name = "apple-service"
            type = "NodePort"
            ports = [{port = 5678}]
        }
    ]
    ingresses = [
        ingress.Ingress {
            name = "example-ingress"
            rules = [
                {
                    http.paths = [
                        {
                            path = "/apple"
                            pathType = "Prefix"
                            backend.service = {
                                name = "app-service"
                                port = {number = 5678}
                            }
                        }
                    ]
                }
            ]
        }
    ]
}
