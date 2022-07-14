# short comment
"""
doc string
"""

schema Container:
    name: str = "main"
    command?: [str]
    ports?: [ContainerPort]
