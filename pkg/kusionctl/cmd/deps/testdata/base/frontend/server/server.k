import base.frontend.container

schema Server:
    name: str
    type: "Deployment" | "StatefulSet" | "DaemonSet" = "Deployment"
    mainContainer: container.Container
    sideCars?: [container.Container]
