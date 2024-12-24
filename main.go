package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	hcAppID  = "9B75AC9EC512F36C84256AC47D813E2C1DD0D6520DF774B020E1E6E2EB29B1F3"
	hcScopes = []string{
		"ReadAccount",
		"Settings",
		"IdentifyAppliance",
		"Control",
		"DeleteAppliance",
		"WriteAppliance",
		"ReadOrigApi",
		"Monitor",
		"WriteOrigApi",
		"Images",
	}
	hcAPIBaseURLs = map[string]string{
		"EU": "https://api.home-connect.com/security/oauth/",
		"NA": "https://api-rna.home-connect.com",
		"CN": "https://api.home-connect.cn",
		"RU": "https://api-rus.home-connect.com",
	}
	hcRedirectURI   = "https://app.home-connect.com/auth/prod"
	hcRedirectURLRE = regexp.MustCompile(".*://.*/auth/prod.*")
	hcAssetBaseURLs = map[string]string{
		"EU": "https://prod.reu.rest.homeconnectegw.com",
		"NA": "https://prod.rna.rest.homeconnectegw.com",
		"CN": "https://prod.rgc.rest.homeconnectegw.cn",
		"RU": "https://prod.rus.rest.homeconnectegw.com",
	}
	selectedRegion = "EU"
)

type deviceProfile struct {
	HCID                      string    `json:"haId"`
	Type                      string    `json:"type"`
	SerialNumber              string    `json:"serialNumber"`
	FeatureMappingFileName    string    `json:"featureMappingFileName"`
	DeviceDescriptionFileName string    `json:"deviceDescriptionFileName"`
	Created                   time.Time `json:"created"`
	ConnectionType            string    `json:"connectionType"`
	Key                       string    `json:"key"`
	IV                        string    `json:"iv,omitempty"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = os.MkdirAll("./output", 0600)

	// retrieve the auth token
	accessToken, err := performInteractiveAuth(ctx)
	if err != nil {
		log.Fatalf("Failed to perform interactive auth: %v", err)
	}
	appliances, err := retrieveAppliances(ctx, accessToken)
	if err != nil {
		log.Fatalf("Failed to retrieve appliances: %v", err)
	}

	profiles := []deviceProfile{}
	for _, appliance := range appliances.Data.HomeAppliances {
		log.Printf("Processing appliance %s (type: %s, serial number: %s)", appliance.Identifier, appliance.Type, appliance.SerialNumber)

		profile := deviceProfile{
			HCID:                      appliance.Identifier,
			Type:                      appliance.Type,
			SerialNumber:              appliance.SerialNumber,
			FeatureMappingFileName:    appliance.Identifier + "_FeatureMapping.xml",
			DeviceDescriptionFileName: appliance.Identifier + "_DeviceDescription.xml",
			Created:                   time.Now(),
		}

		if appliance.TLS != nil {
			log.Printf("Appliance %s is using TLS key %s", appliance.Identifier, appliance.TLS.Key)
			profile.ConnectionType = "TLS"
			profile.Key = appliance.TLS.Key
		} else if appliance.AES != nil {
			log.Printf("Appliance %s is using AES key \"%s\" IV \"%s\"", appliance.Identifier, appliance.AES.Key, appliance.AES.IV)
			profile.ConnectionType = "AES"
			profile.Key = appliance.AES.Key
			profile.IV = appliance.AES.IV
		} else {
			log.Printf("Appliance %s is missing keys, skipping...", appliance.Identifier)
			continue
		}

		profiles = append(profiles, profile)

		profileZIPBytes, err := retrieveDeviceZIP(ctx, accessToken, profile)
		if err != nil {
			log.Fatalf("Failed to generate device ZIP for appliance %s: %v", appliance.Identifier, err)
		}
		log.Printf("Retrieved device ZIP for appliance %s", appliance.Identifier)

		profileZIPPath := filepath.Join("./output", profile.HCID+"_"+profile.Created.Format(time.RFC3339Nano)+".zip")
		if err := os.WriteFile(
			profileZIPPath,
			profileZIPBytes,
			0600,
		); err != nil {
			log.Fatalf("Failed to write ZIP file %s: %w", profileZIPPath, err)
		}
		log.Printf("Saved device %s ZIP file to %s", appliance.Identifier, profileZIPPath)
	}

	// save all profiles to a json file
	profilesBytes, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		log.Fatalf("Failed to encode profiles bundle: %v", err)
	}
	profilesPath := filepath.Join("./output", "profiles.json")
	if err := os.WriteFile(profilesPath, profilesBytes, 0600); err != nil {
		log.Fatalf("Failed to write profiles to %s: %v", profilesPath, err)
	}
	log.Printf("Saved profiles JSON to %s - done!", profilesPath)

	time.Sleep(5 * time.Second)
}
