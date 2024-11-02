package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bartekpacia/fhome/api"
)

// createFhomeClient returns a client that is ready to use.
func createFhomeClient() (*api.Client, error) {
	email := os.Getenv("FHOME_EMAIL")
	if email == "" {
		return nil, fmt.Errorf("FHOME_EMAIL is empty")
	}
	cloudPassword := os.Getenv("FHOME_CLOUD_PASSWORD")
	if cloudPassword == "" {
		return nil, fmt.Errorf("FHOME_CLOUD_PASSWORD is empty")
	}
	resourcePassword := os.Getenv("FHOME_RESOURCE_PASSWORD")
	if resourcePassword == "" {
		return nil, fmt.Errorf("FHOME_RESOURCE_PASSWORD is empty")
	}

	client, err := api.NewClient(nil)
	if err != nil {
		slog.Error("failed to create API client", slog.Any("error", err))
		return nil, err
	} else {
		slog.Debug("created API client")
	}

	err = client.OpenCloudSession(email, cloudPassword)
	if err != nil {
		slog.Error("failed to open client session", slog.Any("error", err))
		return nil, err
	} else {
		slog.Debug("opened client session", slog.String("email", email))
	}

	myResources, err := client.GetMyResources()
	if err != nil {
		slog.Error("failed to get resource", slog.Any("error", err))
		return nil, err
	} else {
		slog.Debug("got resource",
			slog.String("name", myResources.FriendlyName0),
			slog.String("id", myResources.UniqueID0),
			slog.String("type", myResources.ResourceType0),
		)
	}

	err = client.OpenResourceSession(resourcePassword)
	if err != nil {
		slog.Error("failed to open client to resource session", slog.Any("error", err))
		return nil, err
	} else {
		slog.Debug("opened client to resource session")
	}

	return client, nil
}
