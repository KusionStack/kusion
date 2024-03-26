schema App:
    name: str
    idc: str
    cluster: str
    zone: str
    replicas: int

app = App {
    name: "testapp"
    idc: option("idc")
    cluster: option("cluster")
    zone: option("zone")
    replicas: option("replicas")
}