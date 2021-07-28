// Copyright (c) 2020 Oracle and/or its affiliates,  All rights reserved.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package test

type Inputs struct {
	Version  string `yaml:"version"`
	Kind     string `yaml:"kind"`
	Provider string `yaml:"provider"`
	OciTests    []struct {
		Name string            `yaml:"name"`
		Args map[string]string `yaml:","`
	} `yaml:"ocitests"`
}
