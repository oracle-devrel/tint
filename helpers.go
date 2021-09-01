// Copyright (c) 2021 Oracle and/or its affiliates,  All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"

	"github.com/stretchr/testify/assert"
)

func getPublicIp(args map[string]interface{}, terraOut map[string]interface{}) string {

	var ip string
	if strings.Contains(args["public_ip"].(string), "output") {
		prop := strings.Split(args["public_ip"].(string), ".")
		if len(prop) == 2 {
			Println("Retrieving public IP of self..", prop, terraOut[prop[1]], ip)
			ip = terraOut[prop[1]].(string)
		} else {
			Println("Retrieving public IP of another resource..")
			var terraOut1 interface{}
			terraOut1 = fetchAttrDetails(terraOutMap[prop[1]], terraOut1, prop[2], prop[1])
			ip = terraOut1.(map[string]interface{})["public_ip"].(string)
		}
	} else {
		ip = args["public_ip"].(string)
	}
	fmt.Println("Public IP: ", ip)
	return ip
}

func getPrivateIp(t *testing.T, args map[string]interface{}, terraOut map[string]interface{}) []string {

	var priv_ip []string
	priv_ip = strings.Split(args["private_ip"].(string), ",")
	var npriv_ip []string

	for ip := range priv_ip {

		if strings.Contains(priv_ip[ip], "output") {
			prop := strings.Split(priv_ip[ip], ".")

			if len(prop) == 2 {
				npriv_ip = append(npriv_ip, terraOut[prop[1]].(string))
				Println("Retrieving private IP of self..")
			} else {

				Println("Retrieving private IP of another resource..")
				var terraOut1 interface{}
				terraOut1 = fetchAttrDetails(terraOutMap[prop[1]], terraOut1, prop[2], prop[1])
				npriv_ip = append(npriv_ip, terraOut1.(map[string]interface{})["private_ip"].(string))
			}

		} else {
			npriv_ip = append(npriv_ip, priv_ip[ip])
		}

	}

	fmt.Println("Private IPs: ", npriv_ip)
	return npriv_ip
}

func getLbPublicIp(args map[string]interface{}, terraOut map[string]interface{}) string {

	var ip string
	//Split the url http://ip_address
	if strings.Contains(args["url"].(string), "output") {
		prp := strings.Split(args["url"].(string), "/")
		prop := strings.Split(prp[2], ".")

		if len(prop) == 2 {
			if strings.Contains(prop[1], ":") {
				prp2 := strings.Split(prop[1], ":")
				Println("Retrieving public IP of self..", prp2[0], terraOut[prp2[0]], ip)
				//The lb return the ip in a list, so we need to access the element and then convert to string.
				ip = terraOut[prp2[0]].([]interface{})[0].(string) + ":" + prp2[1]
			} else {
				Println("Retrieving public IP of self..", prop, terraOut[prop[1]], ip)
				//The lb return the ip in a list, so we need to access the element and then convert to string.
				ip = terraOut[prop[1]].([]interface{})[0].(string)
			}
		} else {
			Println("Retrieving public IP of another resource..")
			var terraOut1 interface{}
			terraOut1 = fetchAttrDetails(terraOutMap[prop[1]], terraOut1, prop[2], prop[1])
			ip = terraOut1.(map[string]interface{})["ip_addresses"].([]interface{})[0].(string)
		}
	} else {
		prp := strings.Split(args["url"].(string), "//")
		ip = prp[1]
	}
	fmt.Println("Public IP: ", ip)
	return ip
}

func getSshKeyPairs(pub string, priv string, key_pair *ssh.KeyPair) {

	if strings.Contains(pub, "env") && strings.Contains(priv, "env") {

		pub_key, _ := ioutil.ReadFile(os.Getenv(strings.Split(pub, ".")[1]))
		priv_key, _ := ioutil.ReadFile(os.Getenv(strings.Split(priv, ".")[1]))
		key_pair.PublicKey = string(pub_key)
		key_pair.PrivateKey = string(priv_key)

	} else {
		fmt.Println("Invalid input for public or private key, Exiting !!")
		os.Exit(1)
	}
}

func getHost(host *ssh.Host, ip string, u_name string, key_pair *ssh.KeyPair) {

	host.Hostname = ip
	host.SshUserName = u_name
	host.SshKeyPair = key_pair
}

func CheckAndGetOutProp(v interface{}) string {
	// This will output
	var v1 string
	if reflect.TypeOf(v) == reflect.TypeOf(v1) {
		return v1
	}

	_, hasKey := v.(map[interface{}]interface{})["output"]
	if hasKey {
		return v.(map[interface{}]interface{})["output"].(string)
	} else {
		log.Println("Unable to find output property in the yaml file, Exiting..", v)
		Println("Missing Output feild in yaml, Exiting !!")
		os.Exit(1)
	}

	return v.(map[interface{}]interface{})["output"].(string)
}

func fetchAttrDetails(terraOutMap interface{}, terraOut interface{}, attr string, outProp string) interface{} {
    
	//var dummy_string string
	log.Println("Beggining fetchattributes, type of terraOutMap: ",reflect.TypeOf(terraOutMap))
    //os.Exit(1)
	dummy_string := make(map[string]interface{})
	if reflect.TypeOf(terraOutMap) == reflect.TypeOf(dummy_string) {
		if _, ok := terraOutMap.(map[string]interface{})[attr]; ok {
            //log.Printf("Do nothing..")
			return terraOutMap.(map[string]interface{})[attr]
		}else{
            return terraOutMap
		}
	}

	_, hasKey0 := terraOutMap.([]interface{})
	if hasKey0 {
		log.Printf("No attributes to look for, found the data at level 0..")
		for _, v := range terraOutMap.([]interface{}) {
			terraOut = v
		}
		return terraOut
	}

	_, hasKey := terraOutMap.(map[string]interface{})[attr]
	if hasKey {
		log.Printf("Found %v at first level deep..", attr)
		terraOut = terraOutMap.(map[string]interface{})

	} else {
		log.Printf("checking %v at second level deep..", attr)

		for key2, val2 := range terraOutMap.(map[string]interface{}) {

			dummy2 := make(map[string]interface{})
			if reflect.TypeOf(val2) == reflect.TypeOf(dummy2) {
				_, hasKey := val2.(map[string]interface{})[attr]

				if hasKey {
					terraOut = val2.(map[string]interface{})
					log.Printf("Found %v at second level deep..", attr)
				} else {
					log.Printf("Checking %v at third level deep..", attr)
					for _, val3 := range val2.(map[string]interface{}) {
						dummy3 := make(map[string]interface{})
						if reflect.TypeOf(val3) == reflect.TypeOf(dummy3) {
							_, hasKey := val3.(map[string]interface{})[attr]

							if hasKey {
								terraOut = val3.(map[string]interface{})
								log.Printf("Found %v at Third level deep..", attr)
							} else {
								fmt.Printf("Invalid terraform output\nCouldn't find %v from output %v\nThe output may not be conforming to best practices, see an example below: %v", attr, outProp, bestPra)
								os.Exit(1)
							}
						}
					}
				}
			}

			var dummy1 string
			if reflect.TypeOf(val2) == reflect.TypeOf(dummy1) {
				log.Println("FYI, found a simple string in the map  outproperty: ", outProp, "attribute:", attr, "property:", key2)
				continue
			}
		}

		if terraOut != nil {
			_, hasKey := terraOut.(map[string]interface{})[attr]
			if hasKey {
				log.Println("!!!! Found the attribute ..", attr)
			}
		} else {
			fmt.Printf("Invalid terraform output\nCouldn't find %v from output %v\nThe output may not be conforming to best practices, see an example below: %v", attr, outProp, bestPra)
			os.Exit(1)
		}
	}
	return terraOut.(map[string]interface{})[attr]
}

func doTerraChecks(t *testing.T, terraOut interface{}, attr string, checks map[interface{}]interface{}) {

	log.Println("Beggining doTerraChecks, Type of terraoutput", reflect.TypeOf(terraOut), terraOut)
	Println("Test Name:: ", attr)
	//Println("Terra ptil", terraOut)
	// 3. Perform the checks now ..
	//1. If the value we have is a simple string then this check logic applies.
	var v3 string
	v2 := make(map[string]interface{})
	var v4 []interface{}
	var curState interface{}

	if reflect.TypeOf(terraOut) == reflect.TypeOf(v3) {
		fmt.Println("dochecks: found a string", terraOut)
		for prop, val := range checks {
			prop := prop.(string)
			Println("  - Requested check:: ", prop, ":", val, "  - Current state :: ", terraOut)
			assert.Equal(t, val.(string), terraOut.(string), "Test case:: %s :: failed !!", attr)

		}
	} else if (reflect.TypeOf(terraOut) == reflect.TypeOf(v2)) {
		//1. If the value we fetched is a map then this check logic applies.
		terraOut = terraOut.(map[string]interface{})
		for prop, val := range checks {
			prop := prop.(string)
			Println("  - Requested check:: ", prop, ":", val)
			curState = terraOut.(map[string]interface{})[prop]
			log.Printf("checking if %v is a map? %v\n", curState, reflect.ValueOf(curState).Kind() == reflect.Map)
			var dummy []interface{}
			//fmt.Println(reflect.ValueOf(v1).Kind())
			if reflect.ValueOf(curState).Kind() == reflect.ValueOf(dummy).Kind() {
				log.Println("Found a []interface{} in current state (ex: security lists checks []interface{}), processing further ..")
				//convert the map value in yaml from map[interface{}]interface{} to map[string]interface{}
				expYamlMap := make(map[string]interface{})
				trimmedcurState := make(map[string]interface{})
				for v1, v2 := range val.(map[interface{}]interface{}) {
					expYamlMap[v1.(string)] = v2
				}
				//trim the current state also to map[string]interface{}
				for c1, c2 := range curState.([]interface{})[0].(map[string]interface{}) {
					trimmedcurState[c1] = c2
				}
				assert.Equal(t, expYamlMap, trimmedcurState, "Test case:: %s :: failed !!", attr)
				continue
			}

			if reflect.ValueOf(curState).Kind() == reflect.Map {
				log.Println("Found a map in the current state (ex: freeform_tags), processing further ..")
				//convert the map value in yaml from map[interface{}]interface{} to map[string]interface{}
				expYamlMap := make(map[string]interface{})
				for v1, v2 := range val.(map[interface{}]interface{}) {
					expYamlMap[v1.(string)] = v2
				}
				assert.Equal(t, expYamlMap, curState, "Test case:: %s :: failed !!", attr)
				continue
			}
			Println("  - Current state :: ", curState)
			assert.Equal(t, val, curState, "Test case:: %s :: failed !!", attr)
		}
	} else if (reflect.TypeOf(terraOut) == reflect.TypeOf(v4) ){
		log.Println("dochecks: found a []interface", terraOut)
		for _, v := range terraOut.([]interface{}) {
			terraOut = v
		}
		log.Println("doChecks: After converting the terraout []interface now the type is: , type is",reflect.TypeOf(terraOut),terraOut)
		var curState interface{}
		for prop, val := range checks {
			prop := prop.(string)
			curState = terraOut.(map[string]interface{})[prop]
			Println("  - Requested check:: ", prop, ":", val, "  - Current state :: ", curState)
			assert.Equal(t, val.(string), curState, "Test case:: %s :: failed !!", attr)

		}
	} else{
		Println("Invalid output Exiting !!")
		os.Exit(1)
	}

}

func configureTerraformOptions(t *testing.T, terraformDir string) *terraform.Options {

	type Inputs1 struct {
	}
	var inputs1 Inputs1
	err, config := GetInputYamlConfig("../test/inputs.yaml", &inputs1)

	Vars1 := make(map[string]interface{})
	_, hasKey := config.(map[interface{}]interface{})
	if hasKey {
		for i, v := range config.(map[interface{}]interface{}) {
			if strings.Contains(v.(string), "env") {
				env_var := strings.Split(v.(string), ".")
				env_val := os.Getenv(env_var[1])
				Vars1[i.(string)] = env_val
				continue
			}
			Vars1[i.(string)] = v

		}
		if err != nil {
			fmt.Errorf(err.Error())
			os.Exit(1)
		}
	} else {
		log.Println("Found empty ../test/inputs.yaml proceeding..")
	}

	terraformOptions := &terraform.Options{
		TerraformDir: "../",
		Vars:         Vars1,
		Upgrade:      true,
	}
	return terraformOptions
}
