package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func retrieveDeviceZIP(ctx context.Context, accessToken string, profile deviceProfile) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet, hcAssetBaseURLs[selectedRegion]+"/api/iddf/v1/iddf/"+profile.HCID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build device details request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve device details: %w", err)
	}

	// copy to memory - todo is this step needed?
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve device details: %w", err)
	}

	// parse the zip, creating a new one at the same time
	reader, err := zip.NewReader(bytes.NewReader(respBytes), resp.ContentLength)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare output zip file: %w", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, file := range reader.File {
		fileWriter, err := writer.CreateHeader(&file.FileHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create header for file %s: %w", file.Name, err)
		}
		fileReader, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}
		_, err = io.Copy(fileWriter, fileReader)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file %s: %w", file.Name, err)
		}
		_ = fileReader.Close()
	}

	// inject the profile json
	profileWriter, err := writer.Create(profile.HCID + ".json")
	if err != nil {
		return nil, fmt.Errorf("failed to create the profile json file: %w", err)
	}
	profileEncoder := json.NewEncoder(profileWriter)
	profileEncoder.SetIndent("", "  ")
	if err := profileEncoder.Encode(profile); err != nil {
		return nil, fmt.Errorf("failed to encode profile to output zip file: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to flush the output zip: %w", err)
	}

	// return the modified zip file
	return buf.Bytes(), nil
}
