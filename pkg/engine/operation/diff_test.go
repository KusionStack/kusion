package operation

import (
	"fmt"
	"kusionstack.io/kusion/pkg/engine/manifest"
	"reflect"
	"testing"

	"github.com/gonvenience/ytbx"
	"github.com/stretchr/testify/assert"
	"kusionstack.io/kusion/pkg/engine/states"
	"kusionstack.io/kusion/pkg/kusionctl/cmd/diff"
	"kusionstack.io/kusion/third_party/dyff"
)

func TestDiff(t *testing.T) {
	prior := "{\n  \"attributes\": {\n    \"attach\": false,\n    \"bridge\": \"\",\n    \"capabilities\": [\n\n    ],\n    \"command\": [\n      \"nginx\",\n      \"-g\",\n      \"daemon on;\"\n    ],\n    \"container_logs\": null,\n    \"cpu_set\": \"\",\n    \"cpu_shares\": 0,\n    \"destroy_grace_seconds\": null,\n    \"devices\": [\n\n    ],\n    \"dns\": [\n\n    ],\n    \"dns_opts\": [\n\n    ],\n    \"dns_search\": [\n\n    ],\n    \"domainname\": \"\",\n    \"entrypoint\": [\n      \"/docker-entrypoint.sh\"\n    ],\n    \"env\": [\n\n    ],\n    \"exit_code\": null,\n    \"gateway\": \"172.17.0.1\",\n    \"group_add\": [\n\n    ],\n    \"healthcheck\": [\n\n    ],\n    \"hostname\": \"3691b2061977\",\n    \"id\": \"3691b2061977e80b263485845cc0aec6c2b3f83705e5550e02ed699aac0b0033\",\n    \"image\": \"sha256:4f380adfc10f4cd34f775ae57a17d2835385efd5251d6dfe0f246b0018fb0399\",\n    \"init\": false,\n    \"ip_address\": \"172.17.0.2\",\n    \"ip_prefix_length\": 16,\n    \"ipc_mode\": \"private\",\n    \"labels\": [\n\n    ],\n    \"links\": [\n\n    ],\n    \"log_driver\": \"json-file\",\n    \"log_opts\": {\n    },\n    \"logs\": false,\n    \"max_retry_count\": 0,\n    \"memory\": 0,\n    \"memory_swap\": 0,\n    \"mounts\": [\n\n    ],\n    \"must_run\": true,\n    \"name\": \"tutorial\",\n    \"network_alias\": null,\n    \"network_data\": [\n      {\n        \"gateway\": \"172.17.0.1\",\n        \"global_ipv6_address\": \"\",\n        \"global_ipv6_prefix_length\": 0,\n        \"ip_address\": \"172.17.0.2\",\n        \"ip_prefix_length\": 16,\n        \"ipv6_gateway\": \"\",\n        \"network_name\": \"bridge\"\n      }\n    ],\n    \"network_mode\": \"default\",\n    \"networks\": null,\n    \"networks_advanced\": [\n\n    ],\n    \"pid_mode\": \"\",\n    \"ports\": [\n      {\n        \"external\": 8000,\n        \"internal\": 86,\n        \"ip\": \"0.0.0.0\",\n        \"protocol\": \"tcp\"\n      }\n    ],\n    \"privileged\": false,\n    \"publish_all_ports\": false,\n    \"read_only\": false,\n    \"remove_volumes\": true,\n    \"restart\": \"no\",\n    \"rm\": false,\n    \"security_opts\": [\n\n    ],\n    \"shm_size\": 64,\n    \"start\": true,\n    \"stdin_open\": false,\n    \"sysctls\": {\n    },\n    \"tmpfs\": {\n    },\n    \"tty\": false,\n    \"ulimit\": [\n\n    ],\n    \"upload\": [\n\n    ],\n    \"user\": \"\",\n    \"userns_mode\": \"\",\n    \"volumes\": [\n\n    ],\n    \"working_dir\": \"\"\n  },\n  \"sensitive_attributes\": [\n\n  ],\n  \"private\": {\n    \"schema_version\": 2\n  }\n}"
	plan := "{\n  \"attributes\": {\n    \"attach\": false,\n    \"bridge\": \"\",\n    \"capabilities\": [\n\n    ],\n    \"command\": [\n      \"nginx\",\n      \"-g\",\n      \"daemon on;\"\n    ],\n    \"container_logs\": null,\n    \"cpu_set\": \"\",\n    \"cpu_shares\": 0,\n    \"destroy_grace_seconds\": null,\n    \"devices\": [\n\n    ],\n    \"dns\": [\n\n    ],\n    \"dns_opts\": [\n\n    ],\n    \"dns_search\": [\n\n    ],\n    \"domainname\": \"\",\n    \"entrypoint\": [\n      \"/docker-entrypoint.sh\"\n    ],\n    \"env\": [\n\n    ],\n    \"exit_code\": null,\n    \"gateway\": \"172.17.0.1\",\n    \"group_add\": [\n\n    ],\n    \"healthcheck\": [\n\n    ],\n    \"hostname\": \"3691b2061977\",\n    \"id\": \"3691b2061977e80b263485845cc0aec6c2b3f83705e5550e02ed699aac0b0033\",\n    \"image\": \"sha256:4f380adfc10f4cd34f775ae57a17d2835385efd5251d6dfe0f246b0018fb0399\",\n    \"init\": false,\n    \"ip_address\": \"172.17.0.2\",\n    \"ip_prefix_length\": 16,\n    \"ipc_mode\": \"private\",\n    \"labels\": [\n\n    ],\n    \"links\": [\n\n    ],\n    \"log_driver\": \"json-file\",\n    \"log_opts\": {\n    },\n    \"logs\": false,\n    \"max_retry_count\": 0,\n    \"memory\": 0,\n    \"memory_swap\": 0,\n    \"mounts\": [\n\n    ],\n    \"must_run\": true,\n    \"name\": \"tutorial\",\n    \"network_alias\": null,\n    \"network_data\": [\n      {\n        \"gateway\": \"172.17.0.1\",\n        \"global_ipv6_address\": \"\",\n        \"global_ipv6_prefix_length\": 0,\n        \"ip_address\": \"172.17.0.2\",\n        \"ip_prefix_length\": 16,\n        \"ipv6_gateway\": \"\",\n        \"network_name\": \"bridge\"\n      }\n    ],\n    \"network_mode\": \"default\",\n    \"networks\": null,\n    \"networks_advanced\": [\n\n    ],\n    \"pid_mode\": \"\",\n    \"ports\": [\n      {\n        \"external\": 8000,\n        \"internal\": 86,\n        \"ip\": \"0.0.0.0\",\n        \"protocol\": \"tcp\"\n      }\n    ],\n    \"privileged\": false,\n    \"publish_all_ports\": false,\n    \"read_only\": false,\n    \"remove_volumes\": true,\n    \"restart\": \"no\",\n    \"rm\": false,\n    \"security_opts\": [\n\n    ],\n    \"shm_size\": 64,\n    \"start\": true,\n    \"stdin_open\": false,\n    \"sysctls\": {\n    },\n    \"tmpfs\": {\n    },\n    \"tty\": false,\n    \"ulimit\": [\n\n    ],\n    \"upload\": [\n\n    ],\n    \"user\": \"\",\n    \"userns_mode\": \"\",\n    \"volumes\": [\n\n    ],\n    \"working_dir\": \"\"\n  },\n  \"sensitive_attributes\": [\n\n  ],\n  \"private\": {\n    \"schema_version\": 2\n  }\n}"
	// plan := "attributes:\n  attach: false\n  bridge: ''\n  capabilities: []\n  command:\n  - nginx\n  - \"-g\"\n  - daemon off;\n  container_logs:\n  cpu_set: ''\n  cpu_shares: 0\n  destroy_grace_seconds:\n  devices: []\n  dns: []\n  dns_opts: []\n  dns_search: []\n  domainname: ''\n  entrypoint:\n  - \"/docker-entrypoint.sh\"\n  env: []\n  exit_code:\n  gateway: 172.17.0.1\n  group_add: []\n  healthcheck: []\n  host: []\n  hostname: 3691b2061977\n  id: 3691b2061977e80b263485845cc0aec6c2b3f83705e5550e02ed699aac0b0033\n  image: sha256:4f380adfc10f4cd34f775ae57a17d2835385efd5251d6dfe0f246b0018fb0399\n  init: false\n  ip_address: 172.17.0.2\n  ip_prefix_length: 16\n  ipc_mode: private\n  labels: []\n  links: []\n  log_driver: json-file\n  log_opts: {}\n  logs: false\n  max_retry_count: 0\n  memory: 0\n  memory_swap: 0\n  mounts: []\n  must_run: true\n  name: tutorial\n  network_alias:\n  network_data:\n  - gateway: 172.17.0.1\n    global_ipv6_address: ''\n    global_ipv6_prefix_length: 0\n    ip_address: 172.17.0.2\n    ip_prefix_length: 16\n    ipv6_gateway: ''\n    network_name: bridge\n  network_mode: default\n  networks:\n  networks_advanced: []\n  pid_mode: ''\n  ports:\n  - external: 8000\n    internal: 86\n    ip: 0.0.0.0\n    protocol: tcp\n  privileged: false\n  publish_all_ports: false\n  read_only: false\n  remove_volumes: true\n  restart: 'no'\n  rm: false\n  security_opts: []\n  shm_size: 64\n  start: true\n  stdin_open: false\n  sysctls: {}\n  tmpfs: {}\n  tty: false\n  ulimit: []\n  upload: []\n  user: ''\n  userns_mode: ''\n  volumes: []\n  working_dir: ''\nsensitive_attributes: []\nprivate:\n  schema_version: 2"

	report, err := diff_for_test(prior, plan)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 0, len(report.Diffs))

	prior = "{\n  \"attributes\": {\n    \"attach\": false,\n    \"bridge\": \"\",\n    \"capabilities\": [\n\n    ],\n    \"command\": [\n      \"nginx\",\n      \"-g\",\n      \"daemon on;\"\n    ],\n    \"container_logs\": null,\n    \"cpu_set\": \"\",\n    \"cpu_shares\": 0,\n    \"destroy_grace_seconds\": null,\n    \"devices\": [\n\n    ],\n    \"dns\": [\n\n    ],\n    \"dns_opts\": [\n\n    ],\n    \"dns_search\": [\n\n    ],\n    \"domainname\": \"\",\n    \"entrypoint\": [\n      \"/docker-entrypoint.sh\"\n    ],\n    \"env\": [\n\n    ],\n    \"exit_code\": null,\n    \"gateway\": \"172.17.0.1\",\n    \"group_add\": [\n\n    ],\n    \"healthcheck\": [\n\n    ],\n    \"hostname\": \"3691b2061977\",\n    \"id\": \"3691b2061977e80b263485845cc0aec6c2b3f83705e5550e02ed699aac0b0033\",\n    \"image\": \"sha256:4f380adfc10f4cd34f775ae57a17d2835385efd5251d6dfe0f246b0018fb0399\",\n    \"init\": false,\n    \"ip_address\": \"172.17.0.2\",\n    \"ip_prefix_length\": 16,\n    \"ipc_mode\": \"private\",\n    \"labels\": [\n\n    ],\n    \"links\": [\n\n    ],\n    \"log_driver\": \"json-file\",\n    \"log_opts\": {\n    },\n    \"logs\": false,\n    \"max_retry_count\": 0,\n    \"memory\": 0,\n    \"memory_swap\": 0,\n    \"mounts\": [\n\n    ],\n    \"must_run\": true,\n    \"name\": \"tutorial\",\n    \"network_alias\": null,\n    \"network_data\": [\n      {\n        \"gateway\": \"172.17.0.1\",\n        \"global_ipv6_address\": \"\",\n        \"global_ipv6_prefix_length\": 0,\n        \"ip_address\": \"172.17.0.2\",\n        \"ip_prefix_length\": 16,\n        \"ipv6_gateway\": \"\",\n        \"network_name\": \"bridge\"\n      }\n    ],\n    \"network_mode\": \"default\",\n    \"networks\": null,\n    \"networks_advanced\": [\n\n    ],\n    \"pid_mode\": \"\",\n    \"ports\": [\n      {\n        \"external\": 8000,\n        \"internal\": 86,\n        \"ip\": \"0.0.0.0\",\n        \"protocol\": \"tcp\"\n      }\n    ],\n    \"privileged\": false,\n    \"publish_all_ports\": false,\n    \"read_only\": false,\n    \"remove_volumes\": true,\n    \"restart\": \"no\",\n    \"rm\": false,\n    \"security_opts\": [\n\n    ],\n    \"shm_size\": 64,\n    \"start\": true,\n    \"stdin_open\": false,\n    \"sysctls\": {\n    },\n    \"tmpfs\": {\n    },\n    \"tty\": false,\n    \"ulimit\": [\n\n    ],\n    \"upload\": [\n\n    ],\n    \"user\": \"\",\n    \"userns_mode\": \"\",\n    \"volumes\": [\n\n    ],\n    \"working_dir\": \"\"\n  },\n  \"sensitive_attributes\": [\n\n  ],\n  \"private\": {\n    \"schema_version\": 2\n  }\n}"
	plan = "attributes:\n  attach: false\n  bridge: ''\n  capabilities: []\n  command:\n  - nginx\n  - \"-g\"\n  - daemon off;\n  container_logs:\n  cpu_set: ''\n  cpu_shares: 0\n  destroy_grace_seconds:\n  devices: []\n  dns: []\n  dns_opts: []\n  dns_search: []\n  domainname: ''\n  entrypoint:\n  - \"/docker-entrypoint.sh\"\n  env: []\n  exit_code:\n  gateway: 172.17.0.1\n  group_add: []\n  healthcheck: []\n  host: []\n  hostname: 3691b2061977\n  id: 3691b2061977e80b263485845cc0aec6c2b3f83705e5550e02ed699aac0b0033\n  image: sha256:4f380adfc10f4cd34f775ae57a17d2835385efd5251d6dfe0f246b0018fb0399\n  init: false\n  ip_address: 172.17.0.2\n  ip_prefix_length: 16\n  ipc_mode: private\n  labels: []\n  links: []\n  log_driver: json-file\n  log_opts: {}\n  logs: false\n  max_retry_count: 0\n  memory: 0\n  memory_swap: 0\n  mounts: []\n  must_run: true\n  name: tutorial\n  network_alias:\n  network_data:\n  - gateway: 172.17.0.1\n    global_ipv6_address: ''\n    global_ipv6_prefix_length: 0\n    ip_address: 172.17.0.2\n    ip_prefix_length: 16\n    ipv6_gateway: ''\n    network_name: bridge\n  network_mode: default\n  networks:\n  networks_advanced: []\n  pid_mode: ''\n  ports:\n  - external: 8000\n    internal: 86\n    ip: 0.0.0.0\n    protocol: tcp\n  privileged: false\n  publish_all_ports: false\n  read_only: false\n  remove_volumes: true\n  restart: 'no'\n  rm: false\n  security_opts: []\n  shm_size: 64\n  start: true\n  stdin_open: false\n  sysctls: {}\n  tmpfs: {}\n  tty: false\n  ulimit: []\n  upload: []\n  user: ''\n  userns_mode: ''\n  volumes: []\n  working_dir: ''\nsensitive_attributes: []\nprivate:\n  schema_version: 2"
	report, err = diff_for_test(prior, plan)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(report.Diffs))
}

func TestDiffReport(t *testing.T) {
	prior := "{\n  \"attributes\": {\n    \"attach\": false,\n    \"bridge\": \"\",\n    \"capabilities\": [\n\n    ],\n    \"command\": [\n      \"nginx\",\n      \"-g\",\n      \"daemon on;\"\n    ],\n    \"container_logs\": null,\n    \"cpu_set\": \"\",\n    \"cpu_shares\": 0,\n    \"destroy_grace_seconds\": null,\n    \"devices\": [\n\n    ],\n    \"dns\": [\n\n    ],\n    \"dns_opts\": [\n\n    ],\n    \"dns_search\": [\n\n    ],\n    \"domainname\": \"\",\n    \"entrypoint\": [\n      \"/docker-entrypoint.sh\"\n    ],\n    \"env\": [\n\n    ],\n    \"exit_code\": null,\n    \"gateway\": \"172.17.0.1\",\n    \"group_add\": [\n\n    ],\n    \"healthcheck\": [\n\n    ],\n    \"hostname\": \"3691b2061977\",\n    \"id\": \"3691b2061977e80b263485845cc0aec6c2b3f83705e5550e02ed699aac0b0033\",\n    \"image\": \"sha256:4f380adfc10f4cd34f775ae57a17d2835385efd5251d6dfe0f246b0018fb0399\",\n    \"init\": false,\n    \"ip_address\": \"172.17.0.2\",\n    \"ip_prefix_length\": 16,\n    \"ipc_mode\": \"private\",\n    \"labels\": [\n\n    ],\n    \"links\": [\n\n    ],\n    \"log_driver\": \"json-file\",\n    \"log_opts\": {\n    },\n    \"logs\": false,\n    \"max_retry_count\": 0,\n    \"memory\": 0,\n    \"memory_swap\": 0,\n    \"mounts\": [\n\n    ],\n    \"must_run\": true,\n    \"name\": \"tutorial\",\n    \"network_alias\": null,\n    \"network_data\": [\n      {\n        \"gateway\": \"172.17.0.1\",\n        \"global_ipv6_address\": \"\",\n        \"global_ipv6_prefix_length\": 0,\n        \"ip_address\": \"172.17.0.2\",\n        \"ip_prefix_length\": 16,\n        \"ipv6_gateway\": \"\",\n        \"network_name\": \"bridge\"\n      }\n    ],\n    \"network_mode\": \"default\",\n    \"networks\": null,\n    \"networks_advanced\": [\n\n    ],\n    \"pid_mode\": \"\",\n    \"ports\": [\n      {\n        \"external\": 8000,\n        \"internal\": 86,\n        \"ip\": \"0.0.0.0\",\n        \"protocol\": \"tcp\"\n      }\n    ],\n    \"privileged\": false,\n    \"publish_all_ports\": false,\n    \"read_only\": false,\n    \"remove_volumes\": true,\n    \"restart\": \"no\",\n    \"rm\": false,\n    \"security_opts\": [\n\n    ],\n    \"shm_size\": 64,\n    \"start\": true,\n    \"stdin_open\": false,\n    \"sysctls\": {\n    },\n    \"tmpfs\": {\n    },\n    \"tty\": false,\n    \"ulimit\": [\n\n    ],\n    \"upload\": [\n\n    ],\n    \"user\": \"\",\n    \"userns_mode\": \"\",\n    \"volumes\": [\n\n    ],\n    \"working_dir\": \"\"\n  },\n  \"sensitive_attributes\": [\n\n  ],\n  \"private\": {\n    \"schema_version\": 2\n  }\n}"
	plan := "attributes:\n  attach: false\n  bridge: ''\n  capabilities: []\n  command:\n  - nginx\n  - \"-g\"\n  - daemon off;\n  container_logs:\n  cpu_set: ''\n  cpu_shares: 0\n  destroy_grace_seconds:\n  devices: []\n  dns: []\n  dns_opts: []\n  dns_search: []\n  domainname: ''\n  entrypoint:\n  - \"/docker-entrypoint.sh\"\n  env: []\n  exit_code:\n  gateway: 172.17.0.1\n  group_add: []\n  healthcheck: []\n  host: []\n  hostname: 3691b2061977\n  id: 3691b2061977e80b263485845cc0aec6c2b3f83705e5550e02ed699aac0b0033\n  image: sha256:4f380adfc10f4cd34f775ae57a17d2835385efd5251d6dfe0f246b0018fb0399\n  init: false\n  ip_address: 172.17.0.2\n  ip_prefix_length: 16\n  ipc_mode: private\n  labels: []\n  links: []\n  log_driver: json-file\n  log_opts: {}\n  logs: false\n  max_retry_count: 0\n  memory: 0\n  memory_swap: 0\n  mounts: []\n  must_run: true\n  name: tutorial\n  network_alias:\n  network_data:\n  - gateway: 172.17.0.1\n    global_ipv6_address: ''\n    global_ipv6_prefix_length: 0\n    ip_address: 172.17.0.2\n    ip_prefix_length: 16\n    ipv6_gateway: ''\n    network_name: bridge\n  network_mode: default\n  networks:\n  networks_advanced: []\n  pid_mode: ''\n  ports:\n  - external: 8000\n    internal: 86\n    ip: 0.0.0.0\n    protocol: tcp\n  privileged: false\n  publish_all_ports: false\n  read_only: false\n  remove_volumes: true\n  restart: 'no'\n  rm: false\n  security_opts: []\n  shm_size: 64\n  start: true\n  stdin_open: false\n  sysctls: {}\n  tmpfs: {}\n  tty: false\n  ulimit: []\n  upload: []\n  user: ''\n  userns_mode: ''\n  volumes: []\n  working_dir: ''\nsensitive_attributes: []\nprivate:\n  schema_version: 2"

	s, err := DiffReport(prior, plan, diff.OutputHuman)
	assert.NoError(t, err)
	fmt.Println(s)
}

func diff_for_test(prior, plan string) (*dyff.Report, error) {
	to, err := LoadFile(prior, "Last State")
	if err != nil {
		return nil, err
	}
	from, err := LoadFile(plan, "Request State")
	if err != nil {
		return nil, err
	}

	report, err := dyff.CompareInputFiles(from, to, dyff.IgnoreOrderChanges(true))
	return &report, err
}

//func TestDiff2ReleaseDiff(t *testing.T) {
//	request := SetUp(t)
//	marshal := jsonUtil.MustMarshal2String(request)
//	println(marshal)
//	diff, err := Diff(request)
//	assert.NoError(t, err)
//	fmt.Println(diff)
//}

func TestDiff_Diff(t *testing.T) {
	type fields struct {
		StateStorage states.StateStorage
	}
	type args struct {
		request *DiffRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			fields: fields{
				StateStorage: nil,
			},
			args: args{
				request: &DiffRequest{},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Diff{
				StateStorage: tt.fields.StateStorage,
			}
			got, err := d.Diff(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Diff.Diff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Diff.Diff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiffWithRequestResourceAndState(t *testing.T) {
	type args struct {
		plan   states.Resources
		latest *states.State
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				plan: states.Resources{},
				latest: &states.State{
					ID:            0,
					Tenant:        "",
					Stack:         "",
					Project:       "",
					Version:       0,
					KusionVersion: "",
					Serial:        0,
					Operator:      "",
					Resources:     []states.ResourceState{},
				},
			},
			want: `       ___ ________
  ____/ (_) __/ __/  between Last State
 / __  / / /_/ /_        and Request State
/ /_/ / / __/ __/
\__,_/_/_/ /_/      returned no differences

`,
			wantErr: false,
		},
		{
			name: "t2",
			args: args{
				plan:   states.Resources{},
				latest: nil,
			},
			want: `       ___ ________
  ____/ (_) __/ __/  between Last State
 / __  / / /_/ /_        and Request State
/ /_/ / / __/ __/
\__,_/_/_/ /_/      returned one difference

(root level)
± type change from map to list
  -
  +

`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiffWithRequestResourceAndState(&manifest.Manifest{Resources: tt.args.plan}, tt.args.latest)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffWithRequestResourceAndState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DiffWithRequestResourceAndState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildReport(t *testing.T) {
	type args struct {
		mode   string
		report dyff.Report
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				mode: "test-mode",
				report: dyff.Report{
					From:  ytbx.InputFile{},
					To:    ytbx.InputFile{},
					Diffs: []dyff.Diff{},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "t2",
			args: args{
				mode: diff.OutputRaw,
				report: dyff.Report{
					From:  ytbx.InputFile{},
					To:    ytbx.InputFile{},
					Diffs: []dyff.Diff{},
				},
			},
			want:    "diffs: []\n",
			wantErr: false,
		},
		{
			name: "t3",
			args: args{
				mode: diff.OutputHuman,
				report: dyff.Report{
					From:  ytbx.InputFile{},
					To:    ytbx.InputFile{},
					Diffs: []dyff.Diff{},
				},
			},
			want: `       ___ ________
  ____/ (_) __/ __/  between
 / __  / / /_/ /_        and
/ /_/ / / __/ __/
\__,_/_/_/ /_/      returned no differences

`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildReport(tt.args.mode, tt.args.report)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("buildReport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadFile(t *testing.T) {
	doc, _ := ytbx.LoadYAMLDocuments([]byte("a: b"))
	type args struct {
		yaml     string
		location string
	}
	tests := []struct {
		name    string
		args    args
		want    ytbx.InputFile
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				yaml: "a: b",
			},
			want: ytbx.InputFile{
				Documents: doc,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFile(tt.args.yaml, tt.args.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeReport(t *testing.T) {
	type args struct {
		report dyff.Report
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				report: dyff.Report{
					From:  ytbx.InputFile{},
					To:    ytbx.InputFile{},
					Diffs: []dyff.Diff{},
				},
			},
			want: `       ___ ________
  ____/ (_) __/ __/  between
 / __  / / /_/ /_        and
/ /_/ / / __/ __/
\__,_/_/_/ /_/      returned no differences

`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := writeReport(tt.args.report)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeReport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("writeReport() = %v, want %v", got, tt.want)
			}
		})
	}
}
