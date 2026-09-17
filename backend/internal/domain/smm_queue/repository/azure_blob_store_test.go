package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"mi-tech/internal/domain/smm_queue/service"
	"mi-tech/internal/shared/config"
)

type settingsMap map[string]string

func (m settingsMap) Get(key string) (string, error) {
	return m[key], nil
}

func TestNewAzureBlobStoreSupportsSASToken(t *testing.T) {
	store, err := NewAzureBlobStoreWithSASToken("", "mptechstg", "sv=2026-01-01&sig=test", "mp-smm-queue")

	require.NoError(t, err)
	require.NotNil(t, store.client)
	require.Equal(t, "mp-smm-queue", store.container)
	require.Contains(t, store.client.URL(), "sv=2026-01-01")
}

func TestSettingsAzureBlobStoreReadsDatabaseCredentialsPerOperation(t *testing.T) {
	settings := config.NewSettingsProvider(settingsMap{
		"azure_storage_account_name": "mptechstg",
		"azure_storage_sas_token":    "?sv=2026-01-01&sig=db-token",
		"smm_queue_container":        "mp-smm-queue",
	})
	store := NewSettingsAzureBlobStore(settings)

	client, container, err := store.clientForRequest()

	require.NoError(t, err)
	require.NotNil(t, client)
	require.Equal(t, "mp-smm-queue", container)
	require.Contains(t, client.URL(), "sig=db-token")
}

func TestSettingsAzureBlobStoreReturnsUnavailableWithoutCredentials(t *testing.T) {
	t.Setenv("AZURE_STORAGE_CONNECTION_STRING", "")
	t.Setenv("AZURE_STORAGE_SAS_TOKEN", "")
	settings := config.NewSettingsProvider(settingsMap{"smm_queue_container": "mp-smm-queue"})
	store := NewSettingsAzureBlobStore(settings)

	err := store.Upload(context.Background(), "folder/file.txt", []byte("data"), "text/plain")

	require.Error(t, err)
	require.True(t, errors.Is(err, service.ErrStorageUnavailable))
}
