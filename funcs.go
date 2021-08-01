//# Copyright (c) 2021 Oracle and/or its affiliates,  All rights reserved.
//# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/stretchr/testify/assert"
)

/*
How to write your function:
1. Use funcs.go and implement your function, There are few implemented functions which can be used as example.
The function should accept 2 params, terraOut and args coming from the main framework.
2. Add your function name to register.go
3. Provide details of your function in a wiki page.
*/

func httpStatusTestPublic(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	expected_status := args["expected_status"].(int)
	time_out_secs := args["time_out_in_secs"].(int)
	ip := getLbPublicIp(args, terraOut)
	prp := strings.Split(args["url"].(string), "/")

	var new_url []string
	new_url = append(new_url, "http:/")
	for i := range prp {
		if prp[i] == "http:" {
			continue
		}
		if strings.Contains(prp[i], "output") {
			new_url = append(new_url, ip)
			continue
		}
		if prp[i] != "" {
			new_url = append(new_url, prp[i])
		}

	}

	maxRetries := time_out_secs / 5
	timeBetweenRetries := 5 * time.Second
	var status_code int

	_, err := retry.DoWithRetryE(t, fmt.Sprintf("HTTP GET to URL %s", strings.Join(new_url, "/")), maxRetries, timeBetweenRetries, func() (string, error) {

		status, err := HTTPGetWithStatusValidationE(t, strings.Join(new_url, "/"), expected_status)
		status_code = status
		return "", err
	})

	if err != nil {
		t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! Reason: %s", args["oci_test_name"], args["oci_generic_test_name"], err.Error())
	}

	if status_code != -1 {
		assert.Equal(t, expected_status, status_code)
	}
}

func sshTestPublic(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	var key_pair ssh.KeyPair
	pub := args["public_key"].(string)
	priv := args["private_key"].(string)
	getSshKeyPairs(pub, priv, &key_pair)
	ip := getPublicIp(args, terraOut)
	u_name := args["user_name"].(string)
	simple_command := "ls -ltr /var/log"

	_, err := SSHToHostE(t, ip, u_name, &key_pair, simple_command)
	if err != nil {
		t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! Reason: %s", args["oci_test_name"], args["oci_generic_test_name"], err.Error())

	}

}

func fileExistsTestPublic(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	var key_pair ssh.KeyPair
	pub := args["public_key"].(string)
	priv := args["private_key"].(string)
	getSshKeyPairs(pub, priv, &key_pair)
	ip := getPublicIp(args, terraOut)

	u_name := args["user_name"].(string)

	file_name := strings.Split(args["file_name"].(string), ",")

	fmt.Printf("Running file exists test, TEST NAME: %s ==> %s", args["oci_test_name"], args["oci_generic_test_name"])
	for i := range file_name {
		if file_name[i] != "" {
			_, err := SSHFileExistsE(t, ip, u_name, &key_pair, file_name[i])
			if err != nil {
				t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! filename: %s Reason: %s ,file may not exist ?", args["oci_test_name"], args["oci_generic_test_name"], file_name[i], err.Error())
			}
		}
	}
}

func RunShellScriptTestPublic(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	var key_pair ssh.KeyPair
	var host *ssh.Host

	pub := args["public_key"].(string)
	priv := args["private_key"].(string)
	getSshKeyPairs(pub, priv, &key_pair)
	ip := getPublicIp(args, terraOut)
	u_name := args["user_name"].(string)

	script := args["script_name"].(string)
	script_data, _ := ioutil.ReadFile(script)
	getSshKeyPairs(pub, priv, &key_pair)
	host = GetHost(ip, u_name, &key_pair)

	time.Sleep(30)

	maxRetries := 10
	timeBetweenRetries := 1 * time.Second
	description := fmt.Sprintf("SCP file to public host %s", ip)

	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {

		err := ssh.ScpFileToE(t, *host, os.FileMode(0777), "/tmp/oci_script.sh", string(script_data))

		if err != nil {
			t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! Reason: %s", args["oci_test_name"], args["oci_generic_test_name"], err.Error())

		}
		return "", err
	})

	retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		out, err := ssh.CheckSshCommandE(t, *host, "'/tmp/oci_script.sh'")

		fmt.Printf("\n ========== \n output Received from the script: %s \n ========== \n", out)

		if err != nil {
			t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! %s", args["oci_test_name"], args["oci_generic_test_name"], err.Error())
		}
		return "", err
	})
}

func pingTestPublic(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	ip := getPublicIp(args, terraOut)
	cmd := "ping"
	opts := []string{"-c 4", "-i 1", string(ip)}
	_, err := RunShellCommand(cmd, opts, 60)
	if err != nil {
		assert.NotContains(t, string(err.Error()), "exit status", "non-zero exit status while running ", cmd)
	}

}

func netcatTestPublic(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	ip := getPublicIp(args, terraOut)
	port := strings.Split(args["port_open"].(string), ",")
	nc_cmd := "nc"

	// Positive test
	fmt.Printf("Running Ports open tests, TEST NAME: %s ==> %s\n", args["oci_test_name"], args["oci_generic_test_name"])
	for i := range port {
		if port[i] != "" {
			opts := []string{"-zv", ip, port[i]}
			_, err := RunShellCommand(nc_cmd, opts, 30)
			if err != nil {
				t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! \n  Port %s not open, Command used: %s %s", args["oci_test_name"], args["oci_generic_test_name"], port[i], nc_cmd, opts)

			}
		}
	}
	// Negative test
	fmt.Printf("Running Ports NOT open tests, TEST NAME: %s ==> %s\n", args["oci_test_name"], args["oci_generic_test_name"])
	port_not_open := strings.Split(args["port_not_open"].(string), ",")
	for i := range port_not_open {
		if port_not_open[i] != "" {
			opts := []string{"-zv", ip, port_not_open[i]}
			_, err := RunShellCommand(nc_cmd, opts, 30)
			if err != nil {
				continue
			}
			t.Helper()
			t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! \n  Port %s should not be open, Command used: %s %s", args["oci_test_name"], args["oci_generic_test_name"], port_not_open[i], nc_cmd, opts)
		}
	}
}

//Private functions below
func sshTestPrivate(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	var key_pair ssh.KeyPair
	pub := args["public_key"].(string)
	priv := args["private_key"].(string)
	getSshKeyPairs(pub, priv, &key_pair)

	pub_ip := getPublicIp(args, terraOut)
	u_name := args["user_name"].(string)
	priv_ip := getPrivateIp(t, args, terraOut)
	simple_command := "ls -ltr /var/log"

	for ip := range priv_ip {
		_, err := SSHToPrivateHostE(t, pub_ip, priv_ip[ip], u_name, &key_pair, simple_command)
		if err != nil {
			t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! Reason: %s", args["oci_test_name"], args["oci_generic_test_name"], err.Error())

		}
	}
}

func fileExistsTestPrivate(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	var key_pair ssh.KeyPair
	pub := args["public_key"].(string)
	priv := args["private_key"].(string)
	getSshKeyPairs(pub, priv, &key_pair)

	pub_ip := getPublicIp(args, terraOut)
	u_name := args["user_name"].(string)
	priv_ip := getPrivateIp(t, args, terraOut)
	file_name := strings.Split(args["file_name"].(string), ",")

	fmt.Printf("Running file exists test, TEST NAME: %s ==> %s", args["oci_test_name"], args["oci_generic_test_name"])
	for i := range file_name {
		if file_name[i] != "" {

			for ip := range priv_ip {
				_, err := SSHFileExistsPrivateE(t, pub_ip, priv_ip[ip], u_name, &key_pair, file_name[i])
				if err != nil {
					t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! filename: %s Reason: %s ,file may not exist ?", args["oci_test_name"], args["oci_generic_test_name"], file_name[i], err.Error())
				}
			}
		}
	}
}

func processRunningTestPrivate(terraOut map[string]interface{}, args map[string]interface{}, t *testing.T) {

	var key_pair ssh.KeyPair
	pub := args["public_key"].(string)
	priv := args["private_key"].(string)
	getSshKeyPairs(pub, priv, &key_pair)

	pub_ip := getPublicIp(args, terraOut)
	process_name := strings.Split(args["process_name"].(string), ",")
	u_name := args["user_name"].(string)
	priv_ip := getPrivateIp(t, args, terraOut)

	for process := range process_name {
		for ip := range priv_ip {
			command := "sudo ls /run/" + process_name[process] + "/" + process_name[process] + ".pid"
			_, err := SSHToPrivateHostE(t, pub_ip, priv_ip[ip], u_name, &key_pair, command)
			if err != nil {
				t.Errorf("Test Name:: %s ==> %s\n Result:: Failed !! Process name: %s Reason: %s , Process may not exist ?", args["oci_test_name"], args["oci_generic_test_name"], process_name[process], err.Error())

			}
		}
	}

}
