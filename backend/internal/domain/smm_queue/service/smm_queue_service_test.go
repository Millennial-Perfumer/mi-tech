package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type recordedUpload struct {
	name        string
	data        []byte
	contentType string
}

type recordingBlobStore struct {
	uploads       []recordedUpload
	readyFolders  []string
	downloaded    map[string][]byte
	contentTypes  map[string]string
	deletedFolder string
	err           error
}

func (s *recordingBlobStore) Upload(_ context.Context, name string, data []byte, contentType string) error {
	s.uploads = append(s.uploads, recordedUpload{name: name, data: data, contentType: contentType})
	return s.err
}

func (s *recordingBlobStore) ListReadyFolders(context.Context) ([]string, error) {
	return s.readyFolders, s.err
}

func (s *recordingBlobStore) Download(_ context.Context, name string) ([]byte, string, error) {
	if s.downloaded == nil {
		return nil, "", s.err
	}
	data, ok := s.downloaded[name]
	if !ok {
		return nil, "", ErrQueueNotFound
	}
	return data, s.contentTypes[name], s.err
}

func (s *recordingBlobStore) DeleteFolder(_ context.Context, folder string) error {
	s.deletedFolder = folder
	return s.err
}

func TestQueueUploadsReadyMarkerLast(t *testing.T) {
	store := &recordingBlobStore{}
	queue := NewSMMQueueService(store)
	queue.now = func() time.Time { return time.Date(2026, 9, 16, 10, 30, 0, 0, time.FixedZone("IST", 19800)) }
	queue.newID = func() string { return "queue-123" }

	result, err := queue.Queue(context.Background(), "New summer carousel", "#summer, perfume", []MediaUpload{
		{Filename: "first.jpg", Data: []byte{0xff, 0xd8, 0xff, 0xe0, 0x01}},
		{Filename: "second.png", Data: []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0x01}},
	})

	require.NoError(t, err)
	require.Equal(t, "20260916T050000Z_queue-123", result.Folder)
	require.Equal(t, 4, len(store.uploads))
	require.Equal(t, result.Folder+"/01-first.jpg", store.uploads[0].name)
	require.Equal(t, result.Folder+"/02-second.png", store.uploads[1].name)
	require.Equal(t, result.Folder+"/manifest.json", store.uploads[2].name)
	require.Equal(t, result.Folder+"/_READY.json", store.uploads[3].name)
	require.Equal(t, "image/jpeg", store.uploads[0].contentType)
	require.Equal(t, "image/png", store.uploads[1].contentType)

	var manifest QueueManifest
	require.NoError(t, json.Unmarshal(store.uploads[2].data, &manifest))
	require.Equal(t, []string{"#summer", "#perfume"}, manifest.Hashtags)
	require.Equal(t, result.ReadyMarker, manifest.ReadyMarker)
	require.Equal(t, "01-first.jpg", manifest.Media[0].FileName)

	var marker ReadyMarker
	require.NoError(t, json.Unmarshal(store.uploads[3].data, &marker))
	require.Equal(t, result.ManifestPath, marker.Manifest)
}

func TestQueueRejectsInvalidRequestBeforeUploading(t *testing.T) {
	store := &recordingBlobStore{}
	queue := NewSMMQueueService(store)

	_, err := queue.Queue(context.Background(), "", "", []MediaUpload{{Filename: "image.jpg", Data: []byte{0xff, 0xd8, 0xff}}})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidQueue))
	require.Empty(t, store.uploads)

	_, err = queue.Queue(context.Background(), "Caption", "", []MediaUpload{{Filename: "image.jpg", Data: []byte("not an image")}})
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidQueue))
	require.Empty(t, store.uploads)
}

func TestQueueDoesNotPublishReadyMarkerAfterStorageFailure(t *testing.T) {
	store := &recordingBlobStore{err: errors.New("storage down")}
	queue := NewSMMQueueService(store)

	_, err := queue.Queue(context.Background(), "Caption", "", []MediaUpload{{Filename: "image.jpg", Data: []byte{0xff, 0xd8, 0xff, 0xe0}}})
	require.Error(t, err)
	require.Len(t, store.uploads, 1)
	require.NotContains(t, store.uploads[0].name, "_READY.json")
}

func TestListReturnsReadyTimestampFoldersOnly(t *testing.T) {
	store := &recordingBlobStore{readyFolders: []string{
		"20260916T050000Z_newer",
		"20260915T050000Z_older",
		"incomplete-folder",
	}}
	queue := NewSMMQueueService(store)

	items, err := queue.List(context.Background())

	require.NoError(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "20260916T050000Z_newer", items[0].Folder)
	require.Equal(t, "queued", items[0].Status)
	require.Equal(t, "20260915T050000Z_older/_READY.json", items[1].ReadyMarker)
}
