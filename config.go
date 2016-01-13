package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/terraform"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/bigquery/v2"
)

// Config is the configuration structure used to instantiate the Google
// provider.
type Config struct {
	Credentials string
	Project     string
	Region      string

	clientBigQuery  *bigquery.Service
}

func (c *Config) loadAndValidate() error {
	var account accountFile
	clientScopes := []string{
		"https://www.googleapis.com/auth/compute",
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/ndev.clouddns.readwrite",
		"https://www.googleapis.com/auth/devstorage.full_control",
	}

	var client *http.Client

	if c.Credentials != "" {
		contents, _, err := pathorcontents.Read(c.Credentials)
		if err != nil {
			return fmt.Errorf("Error loading credentials: %s", err)
		}

		// Assume account_file is a JSON string
		if err := parseJSON(&account, contents); err != nil {
			return fmt.Errorf("Error parsing credentials '%s': %s", contents, err)
		}

		// Get the token for use in our requests
		log.Printf("[INFO] Requesting Google token...")
		log.Printf("[INFO]   -- Email: %s", account.ClientEmail)
		log.Printf("[INFO]   -- Scopes: %s", clientScopes)
		log.Printf("[INFO]   -- Private Key Length: %d", len(account.PrivateKey))

		conf := jwt.Config{
			Email:      account.ClientEmail,
			PrivateKey: []byte(account.PrivateKey),
			Scopes:     clientScopes,
			TokenURL:   "https://accounts.google.com/o/oauth2/token",
		}

		// Initiate an http.Client. The following GET request will be
		// authorized and authenticated on the behalf of
		// your service account.
		client = conf.Client(oauth2.NoContext)

	} else {
		log.Printf("[INFO] Authenticating using DefaultClient")
		err := error(nil)
		client, err = google.DefaultClient(oauth2.NoContext, clientScopes...)
		if err != nil {
			return err
		}
	}

	versionString := terraform.Version
	prerelease := terraform.VersionPrerelease
	if len(prerelease) > 0 {
		versionString = fmt.Sprintf("%s-%s", versionString, prerelease)
	}
	userAgent := fmt.Sprintf(
		"(%s %s) Terraform/%s", runtime.GOOS, runtime.GOARCH, versionString)

	var err error

	log.Printf("[INFO] Instantiating Google BigQuery Client...")
	c.clientBigQuery, err = bigquery.New(client)
	if err != nil {
		return err
	}
	c.clientBigQuery.UserAgent = userAgent

	return nil
}

// accountFile represents the structure of the account file JSON file.
type accountFile struct {
	PrivateKeyId string `json:"private_key_id"`
	PrivateKey   string `json:"private_key"`
	ClientEmail  string `json:"client_email"`
	ClientId     string `json:"client_id"`
}

func parseJSON(result interface{}, contents string) error {
	r := strings.NewReader(contents)
	dec := json.NewDecoder(r)

	return dec.Decode(result)
}
