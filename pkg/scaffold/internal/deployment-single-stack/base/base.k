import base.pkg.kusion_models.kube.frontend
import base.pkg.kusion_models.kube.frontend.service
import base.pkg.kusion_models.kube.frontend.container
import base.pkg.kusion_models.kube.templates.resource as res_tpl

# Application Configuration
appConfiguration: frontend.Server {
    # Main Container Configuration
    mainContainer = container.Main {
        name = "php-redis"
        env = [
            {
                name = "GET_HOSTS_FROM"
                value = "dns"
            }
        ]
        ports = [
            { containerPort = 80 }
        ]
    }
    selector: {
        "tier" = "frontend"
    }
    podMetadata.labels: {
        "tier" = "frontend"
    }
    schedulingStrategy.resource = res_tpl.medium
    services = [
        service.Service {
            name = "{{ .ServiceName }}"
            type = "NodePort"
            ports = [
                {
                    nodePort = {{ .NodePort }}
                    port = 80
                }
            ]
        }
    ]
}