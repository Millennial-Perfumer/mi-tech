package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"mi-tech/internal/domain/smm_queue/service"
)

type handlerBlobStore struct {
	uploads []string
}

func (s *handlerBlobStore) Upload(_ context.Context, name string, _ []byte, _ string) error {
	s.uploads = append(s.uploads, name)
	return nil
}

func (s *handlerBlobStore) ListReadyFolders(context.Context) ([]string, error) { return nil, nil }

func (s *handlerBlobStore) Download(context.Context, string) ([]byte, string, error) {
	return nil, "", nil
}

func (s *handlerBlobStore) DeleteFolder(context.Context, string) error { return nil }

func TestCreateQueueAcceptsCaptionHashtagsAndOrderedMedia(t *testing.T) {
	store := &handlerBlobStore{}
	queueService := service.NewSMMQueueService(store)
	handler := NewSMMQueueHandler(queueService)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("caption", "A carousel caption"))
	require.NoError(t, writer.WriteField("hashtags", "launch, #perfume"))
	first, err := writer.CreateFormFile("media", "first.jpg")
	require.NoError(t, err)
	_, err = first.Write([]byte{0xff, 0xd8, 0xff, 0xe0})
	require.NoError(t, err)
	second, err := writer.CreateFormFile("media", "second.png")
	require.NoError(t, err)
	_, err = second.Write([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/smm-queue", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	handler.CreateQueue(recorder, req)

	require.Equal(t, http.StatusCreated, recorder.Code)
	var response struct {
		Success bool                `json:"success"`
		Queue   service.QueueResult `json:"queue"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.True(t, response.Success)
	require.Equal(t, 4, len(store.uploads))
	require.Contains(t, store.uploads[0], "/01-first.jpg")
	require.Contains(t, store.uploads[1], "/02-second.png")
	require.Contains(t, store.uploads[3], "/_READY.json")
}

func TestCreateQueueRequiresMultipartMedia(t *testing.T) {
	handler := NewSMMQueueHandler(service.NewSMMQueueService(&handlerBlobStore{}))
	req := httptest.NewRequest(http.MethodPost, "/api/smm-queue", bytes.NewBufferString("caption=hello"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()

	handler.CreateQueue(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}
