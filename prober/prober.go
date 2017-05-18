package prober

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	yaml "gopkg.in/yaml.v2"
	httprobe "k8s.io/kubernetes/pkg/probe/http"
	tcprobe "k8s.io/kubernetes/pkg/probe/tcp"
)

type probeConfig struct {
	configType string
	Service    []struct {
		Name     string
		Protocol string
		IP       string
		Port     int
		TimeOut  time.Duration
	}
}

//Service define file structure
type Service []struct {
	Name     string
	Protocol string
	IP       string
	Port     int
	TimeOut  time.Duration
}

type prober struct {
	http   httprobe.HTTPProber
	tcp    tcprobe.TCPProber
	config probeConfig
}

func (c *probeConfig) getConfigType(configFileName string) error {
	reg := `(\w*.$)`
	regex, err := regexp.Compile(reg)
	if err != nil {
		return err
	}
	result := string(regex.Find([]byte(configFileName)))
	if result == "yml" {
		result = "yaml"
	}
	if result != "yaml" && result != "json" {
		return errors.New("please use yaml or json config file")
	}
	c.configType = result
	return nil
}

func (c *probeConfig) readConfig(configFileName string) error {
	var err error
	err = c.getConfigType(configFileName)
	if err != nil {
		return err
	}

	configFile, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}
	c.convertDataToStruct(configFile)
	return nil
}

func (c *probeConfig) convertDataToStruct(configFile []byte) error {
	var err error
	if c.configType == "yaml" {
		err = yaml.Unmarshal(configFile, &c)
	} else if c.configType == "json" {
		err = json.Unmarshal(configFile, &c)
	}
	if err != nil {
		return err
	}
	return nil
}

//Prober init prober
func Prober(configFileName string, port string) error {
	config := newConfig(configFileName)
	p := newProber(config)
	fmt.Println(p)
	return nil
}

func newProber(c probeConfig) *prober {
	p := &prober{
		http:   httprobe.New(),
		tcp:    tcprobe.New(),
		config: c,
	}
	return p
}

func newConfig(configFileName string) probeConfig {
	c := probeConfig{}
	err := c.readConfig(configFileName)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	return c
}

func (p *prober) serveHTTP(port string) {
	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		ok := true
		errMsg := ""

		if ok {
			w.Write([]byte("OK"))
		} else {
			// Send 503
			http.Error(w, errMsg, http.StatusServiceUnavailable)
		}
	})
	http.ListenAndServe(":"+port, nil)
}
