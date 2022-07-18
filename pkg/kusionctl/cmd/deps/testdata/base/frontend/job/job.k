import base.frontend.container

schema Job:
    name: str
    type: "Job" | "CronJob" = "Job"
    mainContainer: container.Container
    sideCars?: [container.Container]
