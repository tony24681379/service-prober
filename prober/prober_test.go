package prober

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestGetConfigType(t *testing.T) {
	tests := []struct {
		configName     string
		expectedResult string
		expectedError  error
	}{
		{"abc.yaml", "yaml", nil},
		{"123.yml", "yaml", nil},
		{"abc123.json", "json", nil},
		{"abc.asd.a", "", errors.New("please use yaml or json config file")},
	}
	for i, tt := range tests {
		c := probeConfig{}
		err := c.getConfigType(tt.configName)
		if err != nil {
			if err.Error() != tt.expectedError.Error() {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
		if c.configType != tt.expectedResult {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedResult, c.configType)
		}
	}
}

func TestConvertDataToStruct(t *testing.T) {
	expectedServics :=
		service{
			tcpService{
				{
					Name:    "casandra",
					IP:      "127.0.0.1",
					Port:    9042,
					TimeOut: time.Duration(15) * time.Second,
				},
			},
			httpService{
				{
					Name:    "mongo",
					URL:     "http://127.0.0.1:27017",
					TimeOut: time.Duration(15) * time.Second,
				},
			},
		}
	tests := []struct {
		expectedConfigType string
		configFile         []byte
		expectedConfig     probeConfig
		expectedError      error
	}{
		{
			"yaml",
			[]byte(`
---
service:
  http:
  - name: mongo
    url: http://127.0.0.1:27017
    timeout: 15s
  tcp:
  - name: casandra
    ip: 127.0.0.1
    port: 9042
    timeout: 15s
`),
			probeConfig{
				"yaml",
				expectedServics,
			},
			nil,
		},
		{
			"json",
			[]byte(`
{
    "service": {
        "tcp": [
            {
                "name": "casandra",
                "ip": "127.0.0.1",
                "port": 9042,
                "timeout": 15000000000
            }
        ],
        "http":[
            {
                "name": "mongo",
                "url": "http://127.0.0.1:27017",
                "timeout": 15000000000
            }
        ]
    }
}
`),
			probeConfig{
				"json",
				expectedServics,
			},
			nil,
		},
	}
	for i, tt := range tests {
		var c = probeConfig{}
		c.configType = tt.expectedConfigType
		err := c.convertDataToStruct(tt.configFile)
		if !reflect.DeepEqual(c, tt.expectedConfig) {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedConfig, c)
		}
		if err != nil {
			if err != tt.expectedError {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
	}
}
