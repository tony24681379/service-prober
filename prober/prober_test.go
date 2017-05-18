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
		result, err := getConfigType(tt.configName)
		if result != tt.expectedResult {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedResult, result)
		}
		if err != nil {
			if err.Error() != tt.expectedError.Error() {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
	}
}

func TestConvertDataToStruct(t *testing.T) {
	expectedService := services{
		Service{
			{
				Name:     "mongo",
				Protocol: "http",
				IP:       "127.0.0.1:27017",
				TimeOut:  time.Duration(15) * time.Second,
			},
			{
				Name:     "casandra",
				Protocol: "tcp",
				IP:       "127.0.0.1:9042",
				TimeOut:  time.Duration(15) * time.Second,
			},
		},
	}
	tests := []struct {
		configType      string
		configFile      []byte
		expectedService services
		expectedError   error
	}{
		{
			"yaml",
			[]byte(`
---
service:
- name: mongo
  protocol: http
  ip: 127.0.0.1:27017
  timeout: 15s
- name: casandra
  protocol: tcp
  ip: 127.0.0.1:9042
  timeout: 15s
`),
			expectedService,
			nil,
		},
		{
			"json",
			[]byte(`
{
	"service": [
		{
			"name": "mongo",
			"protocol": "http",
			"ip": "127.0.0.1:27017",
			"timeout": 15000000000
		},
		{
            "name": "casandra",
            "protocol": "tcp",
            "ip": "127.0.0.1:9042",
            "timeout": 15000000000
        }
	]
}
`),
			expectedService,
			nil,
		},
	}
	for i, tt := range tests {
		var s = services{}
		err := convertDataToStruct(tt.configType, tt.configFile, &s)
		if !reflect.DeepEqual(s, tt.expectedService) {
			t.Errorf("#%d: expected result=%v, get=%v", i, tt.expectedService, s)
		}
		if err != nil {
			if err.Error() != tt.expectedError.Error() {
				t.Errorf("#%d: expected error=%v, get=%v", i, tt.expectedError, err)
			}
		}
	}
}
