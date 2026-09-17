package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"mi-tech/internal/domain/smm_queue/service"
	"mi-tech/internal/shared/config"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)

type AzureBlobStore struct {
	client    *azblob.Client
	container string
	settings  *config.SettingsProvider
}

func NewAzureBlobStore(connectionString, accountName, container string) (*AzureBlobStore, error) {
	return NewAzureBlobStoreWithSASToken(connectionString, accountName, "", container)
}

// NewAzureBlobStoreWithSASToken creates a static store using one of the
// supported Azure credential forms. It is kept for callers that want a fixed
// client, while the server uses NewSettingsAzureBlobStore for live settings.
func NewAzureBlobStoreWithSASToken(connectionString, accountName, sasToken, container string) (*AzureBlobStore, error) {
	container = strings.TrimSpace(container)
	if container == "" {
		return nil, errors.New("Azure Blob container is not configured")
	}

	client, err := newAzureBlobClient(connectionString, accountName, sasToken)
	if err != nil {
		return nil, err
	}
	return &AzureBlobStore{client: client, container: container}, nil
}

// NewSettingsAzureBlobStore creates a store that reads its credentials and
// container from app_configs on each operation. This means an administrator
// can rotate or replace the Azure credential from Settings without restarting
// the API process.
func NewSettingsAzureBlobStore(settings *config.SettingsProvider) *AzureBlobStore {
	return &AzureBlobStore{settings: settings}
}

func newAzureBlobClient(connectionString, accountName, sasToken string) (*azblob.Client, error) {
	switch {
	case strings.TrimSpace(connectionString) != "":
		return azblob.NewClientFromConnectionString(strings.TrimSpace(connectionString), nil)
	case strings.TrimSpace(accountName) != "" && strings.TrimSpace(sasToken) != "":
		endpoint := "https://" + strings.TrimSpace(accountName) + ".blob.core.windows.net/"
		sasToken = strings.TrimPrefix(strings.TrimSpace(sasToken), "?")
		return azblob.NewClientWithNoCredential(endpoint+"?"+sasToken, nil)
	default:
		return nil, errors.New("Azure Blob connection string or SAS token credentials are not configured")
	}
}

func (s *AzureBlobStore) clientForRequest() (*azblob.Client, string, error) {
	if s == nil {
		return nil, "", service.ErrStorageUnavailable
	}
	if s.settings == nil {
		if s.client == nil || strings.TrimSpace(s.container) == "" {
			return nil, "", service.ErrStorageUnavailable
		}
		return s.client, s.container, nil
	}

	container := strings.TrimSpace(s.settings.GetSMMQueueContainer())
	client, err := newAzureBlobClient(
		s.settings.GetAzureStorageConnectionString(),
		s.settings.GetAzureStorageAccountName(),
		s.settings.GetAzureStorageSASToken(),
	)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", service.ErrStorageUnavailable, err)
	}
	return client, container, nil
}

func (s *AzureBlobStore) Upload(ctx context.Context, blobName string, data []byte, contentType string) error {
	client, container, err := s.clientForRequest()
	if err != nil {
		return err
	}
	contentType = strings.TrimSpace(contentType)
	_, err = client.UploadBuffer(ctx, container, blobName, data, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{BlobContentType: &contentType},
	})
	return err
}

func (s *AzureBlobStore) ListReadyFolders(ctx context.Context) ([]string, error) {
	client, container, err := s.clientForRequest()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	pager := client.NewListBlobsFlatPager(container, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		if page.Segment == nil {
			continue
		}
		for _, item := range page.Segment.BlobItems {
			if item == nil || item.Name == nil {
				continue
			}
			name := *item.Name
			if !strings.HasSuffix(name, "/_READY.json") {
				continue
			}
			seen[strings.TrimSuffix(name, "/_READY.json")] = struct{}{}
		}
	}

	folders := make([]string, 0, len(seen))
	for folder := range seen {
		folders = append(folders, folder)
	}
	sort.Strings(folders)
	return folders, nil
}

func (s *AzureBlobStore) Download(ctx context.Context, blobName string) ([]byte, string, error) {
	client, container, err := s.clientForRequest()
	if err != nil {
		return nil, "", err
	}
	response, err := client.DownloadStream(ctx, container, blobName, nil)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}
	contentType := ""
	if response.ContentType != nil {
		contentType = *response.ContentType
	}
	return data, contentType, nil
}

func (s *AzureBlobStore) DeleteFolder(ctx context.Context, folder string) error {
	client, container, err := s.clientForRequest()
	if err != nil {
		return err
	}
	pager := client.NewListBlobsFlatPager(container, nil)
	var firstErr error
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return err
		}
		if page.Segment == nil {
			continue
		}
		for _, item := range page.Segment.BlobItems {
			if item == nil || item.Name == nil {
				continue
			}
			name := *item.Name
			if name != folder && !strings.HasPrefix(name, folder+"/") {
				continue
			}
			if _, err := client.DeleteBlob(ctx, container, name, nil); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
