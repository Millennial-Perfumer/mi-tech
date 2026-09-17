package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

const (
	MaxMediaFiles = 10
	MaxFileBytes  = 10 * 1024 * 1024
	MaxQueueBytes = 50 * 1024 * 1024
	MaxCaptionLen = 5000
	MaxHashtags   = 30
)

var (
	ErrStorageUnavailable = errors.New("SMM queue storage is not configured")
	ErrInvalidQueue       = errors.New("invalid SMM queue request")
	ErrQueueNotFound      = errors.New("SMM queue item not found")
	allowedContentTypes   = map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
		"image/gif":  ".gif",
	}
	unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	queueFolderPattern  = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}Z_[a-zA-Z0-9-]+$`)
)

// BlobStore is the small storage contract needed by the queue. Keeping it
// here lets the service be tested without contacting Azure.
type BlobStore interface {
	Upload(ctx context.Context, blobName string, data []byte, contentType string) error
	ListReadyFolders(ctx context.Context) ([]string, error)
	Download(ctx context.Context, blobName string) ([]byte, string, error)
	DeleteFolder(ctx context.Context, folder string) error
}

type MediaUpload struct {
	Filename    string
	ContentType string
	Data        []byte
}

type ManifestMedia struct {
	Order       int    `json:"order"`
	FileName    string `json:"file_name"`
	BlobPath    string `json:"blob_path"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
}

type QueueManifest struct {
	SchemaVersion int             `json:"schema_version"`
	QueueID       string          `json:"queue_id"`
	Folder        string          `json:"folder"`
	CreatedAt     time.Time       `json:"created_at"`
	Caption       string          `json:"caption"`
	Hashtags      []string        `json:"hashtags"`
	Media         []ManifestMedia `json:"media"`
	ReadyMarker   string          `json:"ready_marker"`
}

type ReadyMarker struct {
	SchemaVersion int       `json:"schema_version"`
	QueueID       string    `json:"queue_id"`
	Folder        string    `json:"folder"`
	Manifest      string    `json:"manifest"`
	CreatedAt     time.Time `json:"created_at"`
}

type QueueResult struct {
	QueueID      string          `json:"queue_id"`
	Folder       string          `json:"folder"`
	CreatedAt    time.Time       `json:"created_at"`
	ManifestPath string          `json:"manifest_path"`
	ReadyMarker  string          `json:"ready_marker"`
	MediaCount   int             `json:"media_count"`
	Media        []ManifestMedia `json:"media"`
}

type QueueFolder struct {
	Folder      string    `json:"folder"`
	ReadyMarker string    `json:"ready_marker"`
	QueuedAt    time.Time `json:"queued_at"`
	Status      string    `json:"status"`
}

type SMMQueueService struct {
	store BlobStore
	now   func() time.Time
	newID func() string
}

func NewSMMQueueService(store BlobStore) *SMMQueueService {
	return &SMMQueueService{
		store: store,
		now:   time.Now,
		newID: uuid.NewString,
	}
}

// Queue uploads media first, then the manifest, and finally a ready marker.
// Consumers should only process folders containing the ready marker.
func (s *SMMQueueService) Queue(ctx context.Context, caption, rawHashtags string, media []MediaUpload) (*QueueResult, error) {
	if s.store == nil {
		return nil, ErrStorageUnavailable
	}

	caption = strings.TrimSpace(caption)
	if caption == "" {
		return nil, fmt.Errorf("%w: caption is required", ErrInvalidQueue)
	}
	if len(caption) > MaxCaptionLen {
		return nil, fmt.Errorf("%w: caption must be %d characters or fewer", ErrInvalidQueue, MaxCaptionLen)
	}
	if len(media) == 0 {
		return nil, fmt.Errorf("%w: at least one image is required", ErrInvalidQueue)
	}
	if len(media) > MaxMediaFiles {
		return nil, fmt.Errorf("%w: no more than %d images can be queued", ErrInvalidQueue, MaxMediaFiles)
	}

	hashtags, err := normalizeHashtags(rawHashtags)
	if err != nil {
		return nil, err
	}

	totalBytes := 0
	for i := range media {
		if len(media[i].Data) == 0 {
			return nil, fmt.Errorf("%w: image %d is empty", ErrInvalidQueue, i+1)
		}
		if len(media[i].Data) > MaxFileBytes {
			return nil, fmt.Errorf("%w: image %d exceeds the %d MB limit", ErrInvalidQueue, i+1, MaxFileBytes/(1024*1024))
		}
		totalBytes += len(media[i].Data)
		if totalBytes > MaxQueueBytes {
			return nil, fmt.Errorf("%w: carousel exceeds the %d MB total limit", ErrInvalidQueue, MaxQueueBytes/(1024*1024))
		}
		media[i].ContentType = detectAllowedContentType(media[i].Data, media[i].ContentType)
		if media[i].ContentType == "" {
			return nil, fmt.Errorf("%w: image %d must be a JPEG, PNG, WebP, or GIF", ErrInvalidQueue, i+1)
		}
	}

	createdAt := s.now().UTC()
	queueID := s.newID()
	folder := createdAt.Format("20060102T150405Z") + "_" + queueID
	manifestPath := folder + "/manifest.json"
	readyMarkerPath := folder + "/_READY.json"
	manifestMedia := make([]ManifestMedia, 0, len(media))

	for i, item := range media {
		fileName := fmt.Sprintf("%02d-%s", i+1, safeFilename(item.Filename, item.ContentType, i+1))
		blobPath := folder + "/" + fileName
		if err := s.store.Upload(ctx, blobPath, item.Data, item.ContentType); err != nil {
			return nil, fmt.Errorf("upload %s: %w", blobPath, err)
		}
		manifestMedia = append(manifestMedia, ManifestMedia{
			Order:       i + 1,
			FileName:    fileName,
			BlobPath:    blobPath,
			ContentType: item.ContentType,
			SizeBytes:   len(item.Data),
		})
	}

	manifest := QueueManifest{
		SchemaVersion: 1,
		QueueID:       queueID,
		Folder:        folder,
		CreatedAt:     createdAt,
		Caption:       caption,
		Hashtags:      hashtags,
		Media:         manifestMedia,
		ReadyMarker:   readyMarkerPath,
	}
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal SMM queue manifest: %w", err)
	}
	if err := s.store.Upload(ctx, manifestPath, manifestData, "application/json"); err != nil {
		return nil, fmt.Errorf("upload SMM queue manifest: %w", err)
	}

	readyMarker := ReadyMarker{
		SchemaVersion: 1,
		QueueID:       queueID,
		Folder:        folder,
		Manifest:      manifestPath,
		CreatedAt:     createdAt,
	}
	readyData, err := json.MarshalIndent(readyMarker, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal SMM queue ready marker: %w", err)
	}
	if err := s.store.Upload(ctx, readyMarkerPath, readyData, "application/json"); err != nil {
		return nil, fmt.Errorf("upload SMM queue ready marker: %w", err)
	}

	return &QueueResult{
		QueueID:      queueID,
		Folder:       folder,
		CreatedAt:    createdAt,
		ManifestPath: manifestPath,
		ReadyMarker:  readyMarkerPath,
		MediaCount:   len(manifestMedia),
		Media:        manifestMedia,
	}, nil
}

// List returns only folders with a ready marker. Azure may contain partial or
// unrelated blobs; those are intentionally not exposed as queue items.
func (s *SMMQueueService) List(ctx context.Context) ([]QueueFolder, error) {
	if s.store == nil {
		return nil, ErrStorageUnavailable
	}
	folders, err := s.store.ListReadyFolders(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]QueueFolder, 0, len(folders))
	for _, folder := range folders {
		if !isValidQueueFolder(folder) {
			continue
		}
		queuedAt, ok := queuedAtFromFolder(folder)
		if !ok {
			continue
		}
		items = append(items, QueueFolder{
			Folder:      folder,
			ReadyMarker: folder + "/_READY.json",
			QueuedAt:    queuedAt,
			Status:      "queued",
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Folder > items[j].Folder })
	return items, nil
}

func (s *SMMQueueService) Get(ctx context.Context, folder string) (*QueueManifest, error) {
	if s.store == nil {
		return nil, ErrStorageUnavailable
	}
	if !isValidQueueFolder(folder) {
		return nil, ErrQueueNotFound
	}
	data, _, err := s.store.Download(ctx, folder+"/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("download SMM queue manifest: %w", err)
	}
	var manifest QueueManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decode SMM queue manifest: %w", err)
	}
	if manifest.Folder != folder {
		return nil, ErrQueueNotFound
	}
	return &manifest, nil
}

func (s *SMMQueueService) GetMedia(ctx context.Context, folder, filename string) ([]byte, string, error) {
	manifest, err := s.Get(ctx, folder)
	if err != nil {
		return nil, "", err
	}
	for _, media := range manifest.Media {
		if media.FileName != filename || media.BlobPath != folder+"/"+filename || path.Base(filename) != filename {
			continue
		}
		return s.store.Download(ctx, media.BlobPath)
	}
	return nil, "", ErrQueueNotFound
}

func (s *SMMQueueService) Delete(ctx context.Context, folder string) error {
	if s.store == nil {
		return ErrStorageUnavailable
	}
	if !isValidQueueFolder(folder) {
		return ErrQueueNotFound
	}
	return s.store.DeleteFolder(ctx, folder)
}

// Update writes the replacement to a new timestamp folder first, then removes
// the old folder. This keeps the old ready item intact if validation or upload
// of the replacement fails.
func (s *SMMQueueService) Update(ctx context.Context, folder, caption, rawHashtags string, media []MediaUpload) (*QueueResult, error) {
	if !isValidQueueFolder(folder) {
		return nil, ErrQueueNotFound
	}
	result, err := s.Queue(ctx, caption, rawHashtags, media)
	if err != nil {
		return nil, err
	}
	if err := s.Delete(ctx, folder); err != nil {
		return nil, fmt.Errorf("replacement queued in %s but old queue item could not be deleted: %w", result.Folder, err)
	}
	return result, nil
}

func isValidQueueFolder(folder string) bool {
	return queueFolderPattern.MatchString(folder)
}

func queuedAtFromFolder(folder string) (time.Time, bool) {
	if len(folder) < len("20060102T150405Z") || !isValidQueueFolder(folder) {
		return time.Time{}, false
	}
	value, err := time.Parse("20060102T150405Z", folder[:16])
	return value.UTC(), err == nil
}

func normalizeHashtags(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return []string{}, nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || unicode.IsSpace(r) })
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		value := strings.TrimLeft(strings.TrimSpace(part), "#")
		if value == "" {
			continue
		}
		value = "#" + value
		if len(value) > 100 {
			return nil, fmt.Errorf("%w: hashtags must be 100 characters or fewer", ErrInvalidQueue)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		if len(result) == MaxHashtags {
			return nil, fmt.Errorf("%w: no more than %d hashtags can be queued", ErrInvalidQueue, MaxHashtags)
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, nil
}

func detectAllowedContentType(data []byte, supplied string) string {
	detectedContentType := ""
	if len(data) > 0 {
		detectedContentType = http.DetectContentType(data[:min(len(data), 512)])
	}
	contentType := strings.TrimSpace(strings.SplitN(supplied, ";", 2)[0])
	if contentType != "" && contentType != "application/octet-stream" && contentType != detectedContentType {
		return ""
	}
	contentType = detectedContentType
	if _, allowed := allowedContentTypes[contentType]; !allowed {
		return ""
	}
	return contentType
}

func safeFilename(filename, contentType string, index int) string {
	base := filepath.Base(strings.TrimSpace(filename))
	ext := strings.ToLower(filepath.Ext(base))
	if allowedExt, ok := allowedContentTypes[contentType]; !ok || ext != allowedExt && !(allowedExt == ".jpg" && ext == ".jpeg") {
		ext = allowedContentTypes[contentType]
	}
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	stem = unsafeFilenameChars.ReplaceAllString(stem, "-")
	stem = strings.Trim(stem, "-")
	if stem == "" {
		stem = fmt.Sprintf("image-%d", index)
	}
	if len(stem) > 80 {
		stem = stem[:80]
	}
	return stem + ext
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
