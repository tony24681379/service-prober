package prober

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type services struct {
	Service []struct {
		Name     string
		Protocol string
		IP       string
		TimeOut  time.Duration
	}
}

//Service define file structure
type Service []struct {
	Name     string
	Protocol string
	IP       string
	TimeOut  time.Duration
}

//Prober init prober
func Prober(configName string) error {
	s := services{}

	configType, err := getConfigType(configName)
	if err != nil {
		return err
	}
	configFile, err := ioutil.ReadFile(configName)
	if err != nil {
		return err
	}

	err = convertDataToStruct(configType, configFile, &s)
	if err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func getConfigType(configName string) (string, error) {
	reg := `(\w*.$)`
	regex, err := regexp.Compile(reg)
	result := string(regex.Find([]byte(configName)))
	if result == "yml" {
		result = "yaml"
	}
	if result != "yaml" && result != "json" {
		return "", errors.New("please use yaml or json config file")
	}
	return result, err
}

func convertDataToStruct(configType string, configFile []byte, s *services) error {
	var err error
	if configType == "yaml" {
		err = yaml.Unmarshal(configFile, &s)
	} else if configType == "json" {
		err = json.Unmarshal(configFile, &s)
	}
	if err != nil {
		return err
	}
	return nil
}
