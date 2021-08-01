// Copyright (c) 2021 Oracle and/or its affiliates,  All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// GetConfig read test data from a json Config file and parse it into configuration, which is a struct defined by test case
func GetConfig(config_path string, configuration interface{}) error {
	raw, err := ioutil.ReadFile(config_path)
	if err != nil {
		return fmt.Errorf("Unable to read from configuration file: %s", err.Error())
	}
	err = json.Unmarshal(raw, &configuration)
	if err != nil {
		return fmt.Errorf("Failed to parse configurations: %s", err.Error())
	}
	return nil
}

// GetConfig read test data from a yaml Config file and parse it into configuration, which is a struct defined by test case
func GetYamlConfig(config_path string, configuration interface{}) (error, map[string]interface{}) {

	raw, err := ioutil.ReadFile(config_path)
	if err != nil {
		fmt.Errorf("GetYamlConfig(ReadFile): Unable to read from configuration file: %s", err.Error())
		panic(err)
	}

	err = yaml.Unmarshal(raw, &configuration)
	if err != nil {
		fmt.Errorf("GetYamlConfig(Unmarshal): Failed to parse configurations: %s", err.Error())
		panic(err)
	}

	yaml_ok, result := validateYaml(configuration)

	if !yaml_ok {
		log.Println("yaml values are not ok, Exiting..")
		os.Exit(1)
	}
	return err, result
}

func validateYaml(configuration interface{}) (bool, map[string]interface{}) {

	result := make(map[string]interface{})
	//provider := configuration.(map[interface{}]interface{})["provider"]
	//kind := configuration.(map[interface{}]interface{})["kind"]
	version := configuration.(map[interface{}]interface{})["version"]
	tintest := configuration.(map[interface{}]interface{})["TinTests"]
	//generictest := configuration.(map[interface{}]interface{})["GenericTest"]

	if (version != nil) && (tintest != nil) {
		result["version"] = version
		result["tintest"] = tintest
	} else {
		log.Println("Missing yaml fields: version, TinTests .. Exiting !!")
		os.Exit(1)
	}
	/* if generictest != nil {
		result["generictest"] = generictest
	} */
	return true, result
}

// GetConfig read test data from a yaml Config file and parse it into configuration, which is a struct defined by test case
func GetInputYamlConfig(config_path string, configuration interface{}) (error, interface{}) {

	raw, err := ioutil.ReadFile(config_path)
	if err != nil {
		fmt.Errorf("GetYamlConfig(ReadFile): Unable to read from configuration file: %s", err.Error())
		//panic(err)
	}

	err = yaml.Unmarshal(raw, &configuration)
	if err != nil {
		fmt.Errorf("GetYamlConfig(Unmarshal): Failed to parse configurations: %s", err.Error())
		//panic(err)
	}

	return err, configuration
}

func setupLogging() {
	log.Println("Setting up logging ..")
	out, err := os.OpenFile("tin_test.log", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(out)
	log.Println("Done: Setting up logging ..")
}

func Println(v ...interface{}) {
	fmt.Println(v...)
	log.Println(v...)
}
