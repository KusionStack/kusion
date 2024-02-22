import catalog.models.schema.v1 as ac
import catalog.models.schema.v1.workload as wl
import catalog.models.schema.v1.workload.container as c
import catalog.models.schema.v1.workload.network as n
import catalog.models.schema.v1.accessories.mysql

{{ .AppName }}: ac.AppConfiguration {
    workload: wl.Service {
        containers: {
            {{ .AppName }}: c.Container {
                image: "{{ .Image }}"
                env: {
                    # The environment variables for mysql connection.
                    "{{ .AppName | ToUpper }}_DB_NAME": "mysql"
                    "{{ .AppName | ToUpper }}_DB_HOST": "$(KUSION_DB_HOST_{{ .AppName | ToUpper }})"
                    "{{ .AppName | ToUpper }}_DB_USER": "$(KUSION_DB_USERNAME_{{ .AppName | ToUpper }})"
                    "{{ .AppName | ToUpper }}_DB_PASSWORD": "$(KUSION_DB_PASSWORD_{{ .AppName | ToUpper }})"
                }
            }
        }
        replicas: 2
        ports: [
            n.Port {
                port: {{ .Port }}
            }
        ]
    }
    database: {
        "{{ .AppName }}": mysql.MySQL {
            type: "local"
            version: "{{ .DBVersion }}"
        }
    }
}
