package prober

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/kubernetes/pkg/probe"
	//	execprobe "k8s.io/kubernetes/pkg/probe/exec"
	httprobe "k8s.io/kubernetes/pkg/probe/http"
	tcprobe "k8s.io/kubernetes/pkg/probe/tcp"
)

type probeConfig struct {
	configType string
	Service    service
}

type service struct {
	//	Exec []execService
	TCP  []tcpService
	HTTP []httpService
}

type execService struct {
	Name    string
	Cmd     []string
	TimeOut time.Duration
}

type tcpService struct {
	Name    string
	IP      string
	Port    int
	TimeOut time.Duration
}

type httpService struct {
	Name    string
	URL     string
	Header  []httpHeader
	TimeOut time.Duration
}

type httpHeader struct {
	Name  string
	Value string
}

type prober struct {
	//	exec       execprobe.ExecProber
	httpProber httprobe.HTTPProber
	tcpProber  tcprobe.TCPProber
	config     probeConfig
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
	err = c.convertDataToStruct(configFile)
	if err != nil {
		return err
	}
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
	for _, config := range c.Service.HTTP {
		_, err := url.Parse(config.URL)
		if err != nil {
			return err
		}
	}
	return nil
}

// Prober init prober
func Prober(configFileName string, port string) error {
	config := newConfig(configFileName)
	p := newProber(config)
	if glog.V(2) {
		glog.Infof("%+v", p)
	}
	p.serveHTTP(port)
	return nil
}

func newProber(c *probeConfig) *prober {
	p := &prober{
		config: *c,
	}
	if len(c.Service.TCP) > 0 {
		p.tcpProber = tcprobe.New()
	}
	if len(c.Service.HTTP) > 0 {
		p.httpProber = httprobe.New()
	}
	return p
}

func newConfig(configFileName string) *probeConfig {
	c := &probeConfig{}
	err := c.readConfig(configFileName)
	if err != nil {
		glog.Fatalln(err)
		panic(err)
	}
	return c
}

func (p *prober) serveHTTP(port string) {
	http.HandleFunc("/liveness", p.liveness)
	glog.Info("serve on port:", port)
	http.ListenAndServe(":"+port, nil)
}

func buildHeader(headerList []httpHeader) http.Header {
	headers := make(http.Header)
	for _, header := range headerList {
		headers[header.Name] = append(headers[header.Name], header.Value)
	}
	return headers
}

func (p *prober) liveness(w http.ResponseWriter, r *http.Request) {
	ok := true
	var (
		errMsgs []string
		health  probe.Result
		output  string
		err     error
	)
	var wg sync.WaitGroup
	for _, config := range p.config.Service.TCP {
		wg.Add(1)
		go func(config tcpService) {
			defer wg.Done()
			health, output, err = p.tcpProber.Probe(config.IP, config.Port, config.TimeOut)

			errMsg := p.handleError(config.Name, health, output, err)

			if errMsg != "" {
				ok = false
				errMsgs = append(errMsgs, errMsg)
			}
		}(config)
	}
	for _, config := range p.config.Service.HTTP {
		wg.Add(1)
		go func(config httpService) {
			defer wg.Done()
			u, _ := url.Parse(config.URL)
			header := buildHeader(config.Header)
			health, output, err = p.httpProber.Probe(u, header, config.TimeOut)

			errMsg := p.handleError(config.Name, health, output, err)

			if errMsg != "" {
				ok = false
				errMsgs = append(errMsgs, errMsg)
			}
		}(config)
	}
	wg.Wait()
	if ok {
		w.Write([]byte("OK"))
	} else {
		// Send 503
		glog.Warning(errMsgs)
		http.Error(w, strings.Join(errMsgs, ""), http.StatusServiceUnavailable)
	}
}

func (p *prober) handleError(configName string, health probe.Result, output string, err error) string {
	errMsg := ""
	if health != probe.Success {
		errMsg += configName + " " + output + "\n"
	}
	if err != nil {
		errMsg += err.Error()
	}
	return errMsg
}
