schema ContainerPort:
    name?: str
    protocol: "TCP" | "UDP" | "SCTP" = "TCP"
    containerPort: int
