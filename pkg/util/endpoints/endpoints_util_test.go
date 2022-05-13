package endpoints

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestSortSubsets(t *testing.T) {
	subsetsJSON := `	
		[
          	{
				"addresses" : [{"ip": "10.12.1.1"}, {"ip": "10.11.2.2"}, {"ip": "10.10.3.3"}],
				"ports" :     [{"name": "a", "port": 8675}, {"name": "d", "port": 309},{"name": "c", "port": 8801}]
          	},
          	{
				"addresses" : [{"ip": "10.22.2.2"}, {"ip": "10.33.3.33"}, {"ip": "10.11.1.1"}],
				"ports" :     [{"name": "x", "port": 867}, {"name": "z", "port": 311},{"name": "y", "port": 8992}]
          	},
          	{
				"addresses" : [{"ip": "10.31.21.9"}, {"ip": "10.21.2.9"}, {"ip": "10.71.2.2"}],
				"ports" :     [{"name": "a", "port": 791}, {"name": "b", "port": 203},{"name": "c", "port": 7911}]
          	}
         ]
    `
	var subsets []v1.EndpointSubset
	err := json.Unmarshal([]byte(subsetsJSON), &subsets)
	if err != nil {
		t.Fatal("TestSortSubsets failed", err)
	}
	SortSubsets(subsets)
	wantJSON := `	
		[
			{
				"addresses" : [{"ip": "10.10.3.3"}, {"ip": "10.11.2.2"}, {"ip": "10.12.1.1"}],
				"ports" :     [{"name": "c", "port": 8801}, {"name": "d", "port": 309}, {"name": "a", "port": 8675}]
			},
			{
				"addresses" : [{"ip": "10.11.1.1"}, {"ip": "10.22.2.2"}, {"ip": "10.33.3.33"}],
				"ports" :     [{"name":"y","port":8992}, {"name":"x","port":867}, {"name":"z","port":311}]
			},
			{
				"addresses" : [{"ip": "10.21.2.9"}, {"ip": "10.31.21.9"}, {"ip": "10.71.2.2"}],
				"ports" :     [{"name": "a", "port": 791}, {"name": "b", "port": 203}, {"name": "c", "port": 7911}]
			}
		]
	`
	var wantSubsets []v1.EndpointSubset
	err = json.Unmarshal([]byte(wantJSON), &wantSubsets)
	if err != nil {
		t.Fatal("TestSortSubsets failed", err)
	}
	assert.Equal(t, wantSubsets, subsets)
}
