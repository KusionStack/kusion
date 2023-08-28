package net

import (
	"testing"
)

func TestParseHostPort(t *testing.T) {
	tests := []struct {
		name          string
		hostport      string
		expectedHost  string
		expectedPort  string
		expectedError bool
	}{
		{
			name:         "valid host",
			hostport:     "kusionstack.io",
			expectedHost: "kusionstack.io",
			expectedPort: "",
		},
		{
			name:         "valid host:port",
			hostport:     "kusionstack.io:1234",
			expectedHost: "kusionstack.io",
			expectedPort: "1234",
		},
		{
			name:         "valid ip",
			hostport:     "1.2.3.4",
			expectedHost: "1.2.3.4",
			expectedPort: "",
		},
		{
			name:         "valid ip:port",
			hostport:     "1.2.3.4:1234",
			expectedHost: "1.2.3.4",
			expectedPort: "1234",
		},
		{
			name:          "invalid port(not a number)",
			hostport:      "kusionstack.io:aaa",
			expectedError: true,
		},
		{
			name:          "invalid port(>65535)",
			hostport:      "kusionstack.io:65536",
			expectedError: true,
		},
		{
			name:          "invalid port(<1)",
			hostport:      "kusionstack.io:-1",
			expectedError: true,
		},
		{
			name:          "invalid host",
			hostport:      "~~kusionstack.io",
			expectedError: true,
		},
		{
			name:          "invalid host:port",
			hostport:      "~~kusionstack.io:1234",
			expectedError: true,
		},
		{
			name:          "invalid ip",
			hostport:      "1..3.4",
			expectedError: true,
		},
	}

	for _, rt := range tests {
		t.Run(rt.name, func(t *testing.T) {
			actualHost, actualPort, actualError := ParseHostPort(rt.hostport)

			if (actualError != nil) && !rt.expectedError {
				t.Errorf("%s unexpected failure: %v", rt.name, actualError)
				return
			} else if (actualError == nil) && rt.expectedError {
				t.Errorf("%s passed when expected to fail", rt.name)
				return
			}

			if actualHost != rt.expectedHost {
				t.Errorf("%s returned invalid host %s, expected %s", rt.name, actualHost, rt.expectedHost)
				return
			}

			if actualPort != rt.expectedPort {
				t.Errorf("%s returned invalid port %s, expected %s", rt.name, actualPort, rt.expectedPort)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name          string
		port          string
		expectedPort  int
		expectedError bool
	}{
		{
			name:         "valid port",
			port:         "1234",
			expectedPort: 1234,
		},
		{
			name:          "invalid port (not a number)",
			port:          "a",
			expectedError: true,
		},
		{
			name:          "invalid port (<1)",
			port:          "-10",
			expectedError: true,
		},
		{
			name:          "invalid port (>65535)",
			port:          "65536",
			expectedError: true,
		},
	}

	for _, rt := range tests {
		t.Run(rt.name, func(t *testing.T) {
			actualPort, actualError := ParsePort(rt.port)

			if (actualError != nil) && !rt.expectedError {
				t.Errorf("%s unexpected failure: %v", rt.name, actualError)
				return
			} else if (actualError == nil) && rt.expectedError {
				t.Errorf("%s passed when expected to fail", rt.name)
				return
			}

			if actualPort != rt.expectedPort {
				t.Errorf("%s returned invalid port %d, expected %d", rt.name, actualPort, rt.expectedPort)
			}
		})
	}
}
