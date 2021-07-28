// Copyright (c) 2020 Oracle and/or its affiliates,  All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestTint(t *testing.T) {

	setupLogging()
	t.Parallel()
	terraformDir = "../"
	terraformOptions := configureTerraformOptions(t, terraformDir)
	var inputs Inputs
	err, config := GetYamlConfig("../test/test_config.yaml", &inputs)
	if err != nil {
		fmt.Errorf(err.Error())
		os.Exit(1)
	}

	Println("Starting the Terraform Integration Tests ...")
	//defer terraform.Destroy(t, terraformOptions)
	//terraform.InitAndApply(t, terraformOptions)
	//terraform.Plan(t, terraformOptions)
	terraform.Apply(t, terraformOptions)

	for i, v := range config {
		if i == "tintest" {
			doIntTest(v.(map[interface{}]interface{}), t, terraformOptions)
		}
	}
	Println("Completed Terraform Integration Tests ...")
}

func doGenericTest(result map[interface{}]interface{}, t *testing.T, terraformOptions *terraform.Options) {
	fmt.Println("Starting the Generic Tests ...")
	//time.Sleep(30)

	//terraOutMap, _ := terraform.OutputAllE(t, terraformOptions)
	//log.Println("map data full: ", terraOutMap)
	//fmt.Println(result)
	//fmt.Printf("%T", result)
	//f := []interface{}{doTest}

	for k, v := range result {

		log.Println("terratest output to fetch: ", v.(map[interface{}]interface{})["output"], " for: ", k.(string))
		outProp := v.(map[interface{}]interface{})["output"].(string)
		checks := v.(map[interface{}]interface{})["checks"].(map[interface{}]interface{})
		//m := make(map[string]func())

		for prp, val := range checks {
			Println(outProp, prp, val)
			fmt.Printf("%T %T", prp, val)
		}
	}
	assert.Equal(t, 1, 1, "Test failed !! ,expected: ", " current state: ")
}

func doIntTest(result map[interface{}]interface{}, t *testing.T, terraformOptions *terraform.Options) {

	Println("Starting the Terratests...")
	time.Sleep(30)

	terraOutMap, _ = terraform.OutputAllE(t, terraformOptions)
	allFuncs := GetGlobalFuncMap()
	log.Println("saving map data: ", terraOutMap)
	var terraOut interface{}

	for k, v := range result {

		outProp := CheckAndGetOutProp(v)
		attr := k.(string)
		Println("\n#######################################")
		log.Println("terratest output to fetch: ", outProp, " for: ", attr)

		/* Start of checks:

				1. Check if terraform output has some data. look at output.tf for ex:
					output "oci_instances" {
		  				description = "Instance"
		  				value       = module.oci_instances
					}
				2. Next fetch the attribute requested. example: $terraform output oci_instances
				   $ terraform output oci_instances
					{
		 			 "instance" = {
		   	 		"apache_http_server1" = {
		     		 "agent_config" = [
		        	{
		          	"is_monitoring_disabled" = false
		        	},
		      			]
					  "availability_domain"
					  .
					  .
					In the above example, the attribute would be apache_http_server1, apache_http_server2 etc..
				3. Fetch current state and expected state and perform checks.
		*/

		if terraOutMap[outProp] == nil {
			Println("Invalid Yaml !!, Could not find the details for output: ", outProp, " Exiting !!")
			Println("provide a valid output to proceed..")
			os.Exit(1)
		}

		//2. Fetch the attr..
		terraOut = fetchAttrDetails(terraOutMap[outProp], terraOut, attr, outProp)

		_, hasCheck := v.(map[interface{}]interface{})["terra_checks"]
		if hasCheck {
			v1 := make(map[interface{}]interface{})
			if reflect.TypeOf(v.(map[interface{}]interface{})["terra_checks"]) == reflect.TypeOf(v1) {
				checks := v.(map[interface{}]interface{})["terra_checks"].(map[interface{}]interface{})
				doTerraChecks(t, terraOut, attr, checks)
			} else {
				log.Println("terrachecks seems empty, continuing..")
				continue
			}
		}

		_, hasGenCheck := v.(map[interface{}]interface{})["generic_checks"]
		if hasGenCheck {
			for test_name, val := range v.(map[interface{}]interface{})["generic_checks"].(map[interface{}]interface{}) {
				Println("    ++++++++++++++++++++++++++++++++++++++++")
				Println("  - Generic Requested Test:: ", test_name)
				Println("    ++++++++++++++++++++++++++++++++++++++++")
				log.Println("  - Generic Requested Test:: ", test_name, ":", val)
				//log.Println(val.(map[interface{}]interface{})["args"])
				argsmap := make(map[string]interface{})
				argsmap["test_name"] = attr
				argsmap["generic_test_name"] = test_name.(string)

				for k, v := range val.(map[interface{}]interface{})["args"].(map[interface{}]interface{}) {
					argsmap[k.(string)] = v
				}
				_, hasfunc := allFuncs[test_name.(string)]
				if hasfunc {
					allFuncs[test_name.(string)](terraOut.(map[string]interface{}), argsmap, t)
				} else {
					fmt.Printf("Invalid test name: %v provided, Exiting !!", test_name.(string))
					os.Exit(1)
				}
			}
		}
		Println("#######################################")
	}
}
