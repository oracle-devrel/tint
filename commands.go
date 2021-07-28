// Copyright (c) 2020 Oracle and/or its affiliates,  All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/retry"

	"github.com/stretchr/testify/assert"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/ssh"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

func GetHost(host_name string, ssh_user_name string, key_pair *ssh.KeyPair) *ssh.Host {
	host := ssh.Host{
		Hostname:    host_name,
		SshKeyPair:  key_pair,
		SshUserName: ssh_user_name,
	}
	return &host
}

func CheckSSHConnection(t *testing.T, host_name string, ssh_user_name string, key_pair *ssh.KeyPair) {

	Println(t, "Check SSH connection to host %s with username %s", host_name, ssh_user_name, *key_pair)
	host := GetHost(host_name, ssh_user_name, key_pair)
	ssh.CheckSshConnection(t, *host)
}

func CheckSSHConnectionWithRetries(t *testing.T, host_name string, ssh_user_name string, key_pair *ssh.KeyPair,
	max_retries int, time_between_retries time.Duration) {
	logger.Logf(t, "Check SSH connection to host %s with username %s", host_name, ssh_user_name)
	host := GetHost(host_name, ssh_user_name, key_pair)
	retry.DoWithRetry(t, "Checking SSH connection", max_retries, time_between_retries, func() (string, error) {
		return "", ssh.CheckSshConnectionE(t, *host)
	})
}

// SSHToHost SSH to host using its public ip and execute the command given in the parameter
func SSHToHost(t *testing.T, host_name string, ssh_user_name string, key_pair *ssh.KeyPair, command string) string {
	logger.Logf(t, "SSH to host %s with username %s", host_name, ssh_user_name)
	host := GetHost(host_name, ssh_user_name, key_pair)
	logger.Logf(t, "Will ssh to %s and run command %s", host_name, command)
	ssh_result := ssh.CheckSshCommand(t, *host, command)

	return ssh_result
}

// SSHToHost SSH to host using its public ip and execute the command given in the parameter
func SSHToHostE(t *testing.T, host_name string, ssh_user_name string, key_pair *ssh.KeyPair, command string) (string, error) {
	logger.Logf(t, "SSH to host %s with username %s", host_name, ssh_user_name)
	host := GetHost(host_name, ssh_user_name, key_pair)
	logger.Logf(t, "Will ssh to %s and run command %s", host_name, command)
	ssh_result, error := ssh.CheckSshCommandE(t, *host, command)

	return ssh_result, error
}

func SSHFileExistsE(t *testing.T, host_name string, ssh_user_name string, key_pair *ssh.KeyPair, filename string) (string, error) {

	cmd := "ls " + filename
	ssh_result, err := SSHToHostE(t, host_name, ssh_user_name, key_pair, cmd)
	return ssh_result, err
}

func SSHFileExistsPrivateE(t *testing.T, public_host_name string, private_host_name string, ssh_user_name string, key_pair *ssh.KeyPair, filename string) (string, error) {

	cmd := "ls " + filename
	ssh_result, err := SSHToPrivateHostE(t, public_host_name, private_host_name, ssh_user_name, key_pair, cmd)
	return ssh_result, err
}

// SSHToPrivateHost SSH to host using its private ip and execute the command given in the parameter
// A public ip is used to hop to the private ip. This is usually used in the master/slave configuration
// where the master is accessible through public ip while the slaves are accessible through master
func SSHToPrivateHostE(t *testing.T, public_host_name string, private_host_name string, ssh_user_name string,
	key_pair *ssh.KeyPair, command string) (string, error) {

	logger.Logf(t, "SSH to private host %s via public host %s", private_host_name, public_host_name)
	public_host := GetHost(public_host_name, ssh_user_name, key_pair)
	private_host := GetHost(private_host_name, ssh_user_name, key_pair)
	logger.Logf(t, "Will ssh to %s through %s and run command %s", private_host_name, public_host_name, command)
	ssh_result, err := ssh.CheckPrivateSshConnectionE(t, *public_host, *private_host, command)

	return ssh_result, err
}

// HTTPGetWithAuth sends GET request with authorization information (username and password) and returns
// status code and response body
func HTTPGetWithAuth(t *testing.T, url string, username string, password string) (int, string) {
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	req.SetBasicAuth(username, password)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(t, err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(t, err.Error())
	}
	return response.StatusCode, string(body)
}

// HTTPGetWithStatusValidation sends HTTP get request to the URL given in the parameter and verify that
// the response status is expected
func HTTPGetWithStatusValidationE(t *testing.T, url string, expected_status int) (int, error) {

	tlsConfig := tls.Config{}
	status, _, error := http_helper.HttpGetE(t, url, &tlsConfig)
	return status, error

}

// HTTPGetWithBodyValidation sends HTTP get request to the URL given in the parameter and verify that the
// response body is expected
func HTTPGetWithBodyValidation(t *testing.T, url string, expected_body string) {
	tlsConfig := tls.Config{}
	_, body := http_helper.HttpGet(t, url, &tlsConfig)
	assert.Equal(t, strings.Compare(body, expected_body), 0)
}

// GetConfig read test data from a json Config file and parse it into configuration, which is a struct defined by test case
func GetConfig1(config_path string, configuration interface{}) error {
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

func GetResourceProperty(t *testing.T, terraform_options *terraform.Options, property_name string, args ...string) string {
	result := terraform.RunTerraformCommand(t, terraform_options, args...)
	lines := strings.Split(result, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.Contains(line, "=") {
			pair := strings.Split(line, "=")
			if strings.Contains(pair[0], property_name) {
				return strings.Split(pair[1], "\"")[1]
			}
		}
	}
	return ""
}

func RunShellCommand(cmd string, args []string, timeout int) ([]byte, error) {

	// Create a new context and add a timeout to it
	Println("Executing the shell command :: ", cmd, args)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel() // The cancel should be deferred so resources are cleaned up

	// Create the command with our context
	cmdxx := exec.CommandContext(ctx, cmd, args...)

	// This time we can simply use Output() to get the result.
	out, err := cmdxx.Output()

	// We want to check the context error to see if the timeout was executed.
	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Command timed out")
	}

	// If there's no context error, we know the command completed (or errored).
	Println("shell command Output:", string(out))
	if err != nil {
		Println("Non-zero exit code while executing shell command: ", err.Error())
	}

	return out, err
}
