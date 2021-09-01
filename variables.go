// Copyright (c) 2021 Oracle and/or its affiliates,  All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

const bestPra = `

Terraform output examples:

Example-1:

output.tf: 

output "instance_ips" {
value =  "<your requested value>"
}
			
test_config.yaml should look like this:

version: v1
TinTests:
  TEST_INSTANCE:
    output: instance_ips
    checks:
	  state: "AVAILABLE"

Note: It looks for TEST_INSTANCE inside the output instance_ips and then fetches the details.
Then does the checks.

`

var terraformDir string
var terraOutMap map[string]interface{}
