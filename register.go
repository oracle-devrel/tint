// Copyright (c) 2020 Oracle and/or its affiliates,  All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

import "testing"

var allFuncs map[string]func(map[string]interface{}, map[string]interface{}, *testing.T)

func GetGlobalFuncMap() map[string]func(map[string]interface{}, map[string]interface{}, *testing.T) {

	allFuncs := map[string]func(map[string]interface{}, map[string]interface{}, *testing.T){
		"netcat_test_public":       netcatTestPublic,
		"ssh_test_public":          sshTestPublic,
		"ping_test_public":         pingTestPublic,
		"http_status_test_public":  httpStatusTestPublic,
		"file_exists_test_public":  fileExistsTestPublic,
		"RunShellScriptTestPublic": RunShellScriptTestPublic,

		"process_running_test_private": processRunningTestPrivate,
		"ssh_test_private":             sshTestPrivate,
		"file_exists_test_private":     fileExistsTestPrivate,
	}

	return allFuncs

}
