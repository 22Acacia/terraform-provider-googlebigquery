package main

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccBigqueryDatasetCreate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBigQueryDatasetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBigQueryDataset,
				Check: resource.ComposeTestCheckFunc(
					testAccBigQueryDatasetExists(
						"googlebigquery_dataset.foobar"),
				),
			},
		},
	})
}

func TestAccBigqueryDatasetSoftDelete(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBigQueryDatasetExistsThenDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccBigQueryDatasetSD,
				Check: resource.ComposeTestCheckFunc(
					testAccBigQueryDatasetExists(
						"googlebigquery_dataset.foobar_sd"),
				),
			},
		},
	})
}


func testAccCheckBigQueryDatasetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "googlebigquery_dataset" {
			continue
		}

		config := testAccProvider.Meta().(*Config)
		ds, _ := config.clientBigQuery.Datasets.Get(config.Project, rs.Primary.Attributes["datasetId"]).Do()
		if ds != nil {
			return fmt.Errorf("Dataset still present")
		}
		
		fmt.Printf("ds: %q", ds)
	}

	return nil
}

func testAccCheckBigQueryDatasetExistsThenDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "googlebigquery_dataset" {
			continue
		}

		config := testAccProvider.Meta().(*Config)
		ds, _ := config.clientBigQuery.Datasets.Get(config.Project, rs.Primary.Attributes["datasetId"]).Do()		
		if ds == nil {
			return fmt.Errorf("Dataset was deleted when it shouldn't have been!")
		}

		err := config.clientBigQuery.Datasets.Delete(config.Project, rs.Primary.Attributes["datasetId"]).Do()		
		if err != nil {
			return fmt.Errorf("Failed to hard delete soft delete target after check.")
		}
	}
	return nil
}

func testAccBigQueryDatasetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		config := testAccProvider.Meta().(*Config)
		_, err := config.clientBigQuery.Datasets.Get(config.Project, rs.Primary.Attributes["datasetId"]).Do()
		if err != nil {
			return fmt.Errorf("BigQuery Dataset not present")
		}

		return nil
	}
}

const testAccBigQueryDataset = `
resource "googlebigquery_dataset" "foobar" {
	datasetId = "foobar"
	friendlyName = "hi"
}`

const testAccBigQueryDatasetSD = `
resource "googlebigquery_dataset" "foobar_sd" {
	datasetId = "foobar"
	friendlyName = "hi"
	softDelete = true
}`