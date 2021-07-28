### Terraform Integration Test(TINT)

Simple configuration based continuous integration testing tool for terraform developers.
Performs rich production grade tests on your infrastructure as code.Makes use of opensource terratest framework and sdks and provides a cohesive, user friendly config based approach to write your test cases.

### How to run the tests:

# Download your terraform solution module, below shows an example of http solution:
```
$ cd <some-dir>
$ git clone git@github.com:oracle-terraform-modules/terraform-oci-tdf-block-storage.git
$ cd <some-dir>/terraform-oci-tdf-block-storage/examples/simple_block_volume
$ git clone git@orahub.oci.oraclecorp.com:devrel/tint.git
```
# Create a folder named test and test_config.yaml file
```
$ mkdir test
$ ls test/test_config.yaml (Refer sample config files on how to write your config file)
https://confluence.oci.oraclecorp.com/display/ATEAM/Terraform+testing+framework#Terraformtestingframework-Sampletest_config.yamlFile

```

# Create inputs.yaml file (Optional)
``` 
$ ls test/inputs.yaml (This is the input to your terraform module, Note: The name has to be inputs.yaml)
 
Example:
 
dhcp-10-191-131-113:test vsnaik$ cat inputs.yaml
 
default_compartment_id: env.TF_VAR_compartment_id
```

# Run the tests:
```$ cd <some-dir>/terraform-oci-tdf-block-storage/examples/simple_block_volume
$ cd tint
$ go test -run TestTint -timeout 30m
```