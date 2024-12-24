package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type appliancesResponse struct {
	Data struct {
		HomeAppliances []struct {
			AES *struct {
				IV  string `json:"iv"`
				Key string `json:"key"`
			} `json:"aes"`
			Brand             string `json:"brand"`
			CommunicationType string `json:"communicationType"`
			ConfigLabels      struct {
				RedLabels []interface{} `json:"redLabels"`
				Version   string        `json:"version"`
			} `json:"configLabels"`
			Country       string `json:"country"`
			Customerindex string `json:"customerindex"`
			DDFRevision   int    `json:"ddfrevision"`
			DDFVersion    int    `json:"ddfversion"`
			Demo          bool   `json:"demo"`
			ENumber       string `json:"enumber"`
			FDString      string `json:"fdstring"`
			Identifier    string `json:"identifier"`
			Info          string `json:"info"`
			Mac           string `json:"mac"`
			Name          string `json:"name"`
			PairingTime   int64  `json:"pairingTime"`
			Quarantined   bool   `json:"quarantined"`
			SerialNumber  string `json:"serialnumber"`
			TLS           *struct {
				Key string `json:"key"`
			} `json:"tls"`
			Type  string `json:"type"`
			Users []struct {
				CellPhone interface{} `json:"cellPhone"`
				Email     string      `json:"email"`
				Firstname string      `json:"firstname"`
				HCID      string      `json:"hcId"`
				Lastname  string      `json:"lastname"`
				Su        bool        `json:"su"`
			} `json:"users"`
			Vib     string `json:"vib"`
			ZNumber string `json:"znumber"`
		} `json:"homeAppliances"`
		LoginType    string `json:"loginType"`
		SmartDevices []struct {
			Identifier  string `json:"identifier"`
			Language    string `json:"language"`
			Name        string `json:"name"`
			Platform    string `json:"platform"`
			PushVersion int    `json:"pushVersion"`
			PushID      string `json:"pushid"`
			TokenValid  bool   `json:"tokenValid"`
			Type        string `json:"type"`
		} `json:"smartDevices"`
		User struct {
			AccountID           string `json:"accountId"`
			CiamID              string `json:"ciamId"`
			Country             string `json:"country"`
			DataCollection      bool   `json:"dataCollection"`
			Email               string `json:"email"`
			FirstName           string `json:"firstname"`
			Language            string `json:"language"`
			LastName            string `json:"lastname"`
			MarketingNewsletter bool   `json:"marketingNewsletter"`
			MarketingPermission bool   `json:"marketingPermission"`
			RegistrationTime    int64  `json:"registrationTime"`
			TrackingEnabled     bool   `json:"trackingEnabled"`
		} `json:"user"`
	} `json:"data"`
	HCID string `json:"hcId"`
}

func retrieveAppliances(ctx context.Context, accessToken string) (*appliancesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hcAssetBaseURLs[selectedRegion]+"/account/details", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build account details request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve appliances: %w", err)
	}
	var result appliancesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode appliances response: %w", err)
	}
	return &result, nil
}
