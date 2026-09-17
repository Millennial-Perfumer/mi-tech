package config

import (
	"os"
	"testing"
)

type mapConfigRepository map[string]string

func (r mapConfigRepository) Get(key string) (string, error) {
	return r[key], nil
}

func TestSMMQueueSettingsPreferDatabaseValues(t *testing.T) {
	t.Setenv("AZURE_STORAGE_CONNECTION_STRING", "env-connection-string")
	t.Setenv("AZURE_STORAGE_ACCOUNT_NAME", "env-account")
	t.Setenv("AZURE_STORAGE_SAS_TOKEN", "env-sas-token")
	t.Setenv("SMM_QUEUE_CONTAINER", "env-container")

	provider := NewSettingsProvider(mapConfigRepository{
		"azure_storage_connection_string": "db-connection-string",
		"azure_storage_account_name":      "db-account",
		"azure_storage_sas_token":         "db-sas-token",
		"smm_queue_container":             "db-container",
	})

	if got := provider.GetAzureStorageConnectionString(); got != "db-connection-string" {
		t.Fatalf("connection string = %q, want database value", got)
	}
	if got := provider.GetAzureStorageAccountName(); got != "db-account" {
		t.Fatalf("account name = %q, want database value", got)
	}
	if got := provider.GetAzureStorageSASToken(); got != "db-sas-token" {
		t.Fatalf("SAS token = %q, want database value", got)
	}
	if got := provider.GetSMMQueueContainer(); got != "db-container" {
		t.Fatalf("container = %q, want database value", got)
	}
}

func TestSMMQueueSettingsFallBackToEnvironmentAndDefaults(t *testing.T) {
	t.Setenv("AZURE_STORAGE_CONNECTION_STRING", "env-connection-string")
	t.Setenv("AZURE_STORAGE_ACCOUNT_NAME", "env-account")
	t.Setenv("AZURE_STORAGE_SAS_TOKEN", "env-sas-token")
	t.Setenv("SMM_QUEUE_CONTAINER", "env-container")

	provider := NewSettingsProvider(mapConfigRepository{})
	if got := provider.GetAzureStorageConnectionString(); got != "env-connection-string" {
		t.Fatalf("connection string = %q, want environment value", got)
	}
	if got := provider.GetAzureStorageAccountName(); got != "env-account" {
		t.Fatalf("account name = %q, want environment value", got)
	}
	if got := provider.GetAzureStorageSASToken(); got != "env-sas-token" {
		t.Fatalf("SAS token = %q, want environment value", got)
	}
	if got := provider.GetSMMQueueContainer(); got != "env-container" {
		t.Fatalf("container = %q, want environment value", got)
	}

	for _, key := range []string{"AZURE_STORAGE_CONNECTION_STRING", "AZURE_STORAGE_ACCOUNT_NAME", "AZURE_STORAGE_SAS_TOKEN", "SMM_QUEUE_CONTAINER"} {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	if got := provider.GetAzureStorageAccountName(); got != "mptechstg" {
		t.Fatalf("default account name = %q, want mptechstg", got)
	}
	if got := provider.GetSMMQueueContainer(); got != "mp-smm-queue" {
		t.Fatalf("default container = %q, want mp-smm-queue", got)
	}
}
