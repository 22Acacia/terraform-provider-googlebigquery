package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBigqueryTableCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBigQueryTableDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBigQueryTable,
				Check: resource.ComposeTestCheckFunc(
					testAccBigQueryTableExists(
						"googlebigquery_table.foobar"),
				),
			},
		},
	})
}

func TestAccBigqueryTableCreateFieldsFile(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBigQueryTableDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBigQueryTableJsonFile,
				Check: resource.ComposeTestCheckFunc(
					testAccBigQueryTableExists(
						"googlebigquery_table.foobar"),
				),
			},
		},
	})
}

func testAccCheckBigQueryTableDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "googlebigquery_table" {
			continue
		}

		config := testAccProvider.Meta().(*Config)
		_, err := config.clientBigQuery.Tables.Get(config.Project, rs.Primary.Attributes["datasetId"], rs.Primary.Attributes["name"]).Do()
		if err == nil {
			return fmt.Errorf("Table still present")
		}
	}

	return nil
}

func testAccBigQueryTableExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		config := testAccProvider.Meta().(*Config)
		_, err := config.clientBigQuery.Tables.Get(config.Project, rs.Primary.Attributes["datasetId"], rs.Primary.Attributes["name"]).Do()
		if err != nil {
			return fmt.Errorf("BigQuery Table not present")
		}

		return nil
	}
}

const testAccBigQueryTable = `
resource "googlebigquery_dataset" "foobar" {
	datasetId = "foobar"
}

resource "googlebigquery_table" "foobar" {
	tableId = "foobar"
	datasetId = "${googlebigquery_dataset.foobar.datasetId}"
	
	schema {
		description = "field"
		mode = "nullable"
		name = "foo"
		type = "string"
	}
	
	schema {
		name = "bar"
		type = "string"
	}
}`

const testAccBigQueryTableJsonFile = `
resource "googlebigquery_dataset" "foobar" {
	datasetId = "foobar"
}

resource "googlebigquery_table" "foobar" {
	tableId = "foobar"
	datasetId = "${googlebigquery_dataset.foobar.datasetId}"

	schemaFile = "./test-fixtures/fake_bigquery_table.json"
}`
