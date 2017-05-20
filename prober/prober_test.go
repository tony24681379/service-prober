package prober

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"k8s.io/kubernetes/pkg/probe"
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
			[]tcpService{
				{
					Name:    "casandra",
					IP:      "127.0.0.1",
					Port:    9042,
					TimeOut: time.Duration(15) * time.Second,
				},
			},
			[]httpService{
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

type fakeHTTPProber struct {
	result probe.Result
	err    error
}

func (p fakeHTTPProber) Probe(url *url.URL, headers http.Header, timeout time.Duration) (probe.Result, string, error) {
	return p.result, "message", p.err
}

type fakeTCPProber struct {
	result probe.Result
	err    error
}

func (p fakeTCPProber) Probe(host string, port int, timeout time.Duration) (probe.Result, string, error) {
	return p.result, "message", p.err
}

func TestLiveness(t *testing.T) {
	tests := []struct {
		probe          *prober
		expectedResult []byte
	}{
		{
			&prober{
				tcpProber:  fakeTCPProber{result: probe.Success},
				httpProber: fakeHTTPProber{result: probe.Success},
				config: probeConfig{
					Service: service{
						[]tcpService{{Name: "casandra"}},
						[]httpService{{Name: "mongo"}},
					},
				},
			},
			[]byte("OK"),
		},
		{
			&prober{
				tcpProber:  fakeTCPProber{result: probe.Success},
				httpProber: fakeHTTPProber{result: probe.Failure},
				config: probeConfig{
					Service: service{
						[]tcpService{{Name: "casandra"}},
						[]httpService{{Name: "mongo"}},
					},
				},
			},
			[]byte("mongo message\n\n"),
		},
		{
			&prober{
				tcpProber:  fakeTCPProber{result: probe.Failure},
				httpProber: fakeHTTPProber{result: probe.Failure},
				config: probeConfig{
					Service: service{
						[]tcpService{{Name: "casandra"}},
						[]httpService{{Name: "mongo"}},
					},
				},
			},
			[]byte("mongo message\ncasandra message\n\n"),
		},
	}

	for i, tt := range tests {
		ts := httptest.NewServer(http.HandlerFunc(tt.probe.liveness))
		defer ts.Close()
		res, err := http.Get(ts.URL)
		if err != nil {
			log.Fatal(err)
		}
		response, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		if !reflect.DeepEqual(response, tt.expectedResult) {
			t.Errorf("#%d: expected error=%s, get=%s", i, tt.expectedResult, response)
		}
	}
}
