package prober

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/kubernetes/pkg/probe"
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
	for _, config := range c.Service {
		_, err := url.Parse(config.IP + ":" + strconv.Itoa(config.Port))
		if err != nil {
			return err
		}
	}

	return nil
}

//Prober init prober
func Prober(configFileName string, port string) error {
	config := newConfig(configFileName)
	p := newProber(config)
	fmt.Println(p)
	p.serveHTTP(port)
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
	http.HandleFunc("/liveness", func(w http.ResponseWriter, r *http.Request) {
		ok := true
		var (
			errMsgs []string
			health  probe.Result
			output  string
			err     error
		)
		for _, config := range p.config.Service {
			errMsg := ""
			if config.Protocol == "tcp" {
				health, output, err = p.tcp.Probe(config.IP, config.Port, config.TimeOut)
			} else if config.Protocol == "http" {
				u, _ := url.Parse(config.IP + ":" + strconv.Itoa(config.Port))
				health, output, err = p.http.Probe(u, nil, config.TimeOut)
			}

			if health != probe.Success {
				errMsg += config.Name + " " + output + "\n"
				ok = false
			}
			if err != nil {
				errMsg += err.Error()
				ok = false
			}
			errMsgs = append(errMsgs, errMsg)
		}

		if ok {
			w.Write([]byte("OK"))
		} else {
			// Send 503
			log.Println(errMsgs)
			http.Error(w, strings.Join(errMsgs, ""), http.StatusServiceUnavailable)
		}
	})
	log.Print("serve on port:", port)
	http.ListenAndServe(":"+port, nil)
}
