import base.pkg.kusion_models.kube.frontend
import base.pkg.kusion_models.kube.frontend.service
import base.pkg.kusion_models.kube.frontend.container
import base.pkg.kusion_models.kube.templates.resource as res_tpl

# Application Configuration
appConfiguration: frontend.Server {
    # Main Container Configuration
    mainContainer = container.Main {
        ports = [
            {containerPort = 80}
        ]
    }
    image = "nginx:1.7.8"
    services = [
        service.Service {
            name = __META_APP_NAME
            type = "NodePort"
            ports = [
                {
                    nodePort = 30201
                    port = 80
                    targetPort = 80
                }
            ]
        }
    ]
}
