package dyff

import (
	"fmt"
	"testing"
)

func TestComparator(t *testing.T) {
	from := "{\"containers\":[{\"resource\":{\"cpu\":{\"cpuSet\":{\"cpuIDs\":[],\"spreadStrategy\":\"sameCoreFirst\"}}},\"name\":\"cafedpreviewserver\"}],\"affinity\":{\"podAntiAffinity\":{\"requiredDuringSchedulingIgnoredDuringExecution\":[{\"labelSelector\":{\"matchExpressions\":[{\"values\":[\"cafedpreviewserver-prepub\"],\"key\":\"sigma.ali/deploy-unit\",\"operator\":\"In\"}]},\"topologyKey\":\"kubernetes.io/hostname\",\"maxCount\":1}]}}}"
	to := "{\"affinity\": {\"podAntiAffinity\": {\"requiredDuringSchedulingIgnoredDuringExecution\": [{\"labelSelector\": {\"matchExpressions\": [{\"key\": \"sigma.ali/deploy-unit\", \"operator\": \"In\", \"values\": [\"cafedpreviewserver-prepub\"]}]}, \"topologyKey\": \"kubernetes.io/hostname\", \"maxCount\": 1}]}}, \"containers\": [{\"name\": \"cafedpreviewserver\", \"resource\": {\"cpu\": {\"cpuSet\": {\"spreadStrategy\": \"sameCoreFirst\", \"cpuIDs\": []}}}}]}"
	fmt.Printf("the result is %t\n", JsonStrComparator(from, to))
}
