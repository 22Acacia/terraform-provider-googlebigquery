package main

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"credentials": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("GOOGLE_CREDENTIALS", nil),
				ValidateFunc: validateCredentials,
			},

			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GOOGLE_PROJECT", nil),
			},

			"region": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GOOGLE_REGION", nil),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"googlebigquery_dataset":               resourceBigQueryDataset(),
			"googlebigquery_table":                 resourceBigQueryTable(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	credentials := d.Get("credentials").(string)
	if credentials == "" {
		credentials = d.Get("account_file").(string)
	}
	config := Config{
		Credentials: credentials,
		Project:     d.Get("project").(string),
		Region:      d.Get("region").(string),
	}

	if err := config.loadAndValidate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func validateAccountFile(v interface{}, k string) (warnings []string, errors []error) {
	if v == nil {
		return
	}

	value := v.(string)

	if value == "" {
		return
	}

	contents, wasPath, err := pathorcontents.Read(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("Error loading Account File: %s", err))
	}
	if wasPath {
		warnings = append(warnings, `account_file was provided as a path instead of 
as file contents. This support will be removed in the future. Please update
your configuration to use ${file("filename.json")} instead.`)
	}

	var account accountFile
	if err := json.Unmarshal([]byte(contents), &account); err != nil {
		errors = append(errors,
			fmt.Errorf("account_file not valid JSON '%s': %s", contents, err))
	}

	return
}

func validateCredentials(v interface{}, k string) (warnings []string, errors []error) {
	if v == nil || v.(string) == "" {
		return
	}
	creds := v.(string)
	var account accountFile
	if err := json.Unmarshal([]byte(creds), &account); err != nil {
		errors = append(errors,
			fmt.Errorf("credentials are not valid JSON '%s': %s", creds, err))
	}

	return
}
