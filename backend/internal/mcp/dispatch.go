package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	maxMCPQueueMediaFiles      = 10
	maxMCPQueueMediaBytes      = 50 << 20
	maxMCPQueueMediaTotalBytes = 100 << 20
)

// MuxExecutor dispatches allowlisted tool calls to internal read and write
// muxes. Requests execute in-process, so no network hops or additional
// authentication are required beyond the machine key validated at transport.
type MuxExecutor struct {
	handler      http.Handler
	writeHandler http.Handler
}

// identityExecutor binds a machine-key identity to an executor. This is used
// by the stdio transport, where there is no HTTP middleware to inject the
// authenticated key into the MCP request context.
type identityExecutor struct {
	inner  Executor
	id     int64
	name   string
	scopes []string
}

// WithMachineIdentity returns an executor that propagates machine-key
// identity into every dispatched request and audit record.
func WithMachineIdentity(inner Executor, id int64, name string, scopes []string) Executor {
	return &identityExecutor{inner: inner, id: id, name: name, scopes: scopes}
}

func (e *identityExecutor) Dispatch(ctx context.Context, tool ToolSpec, args map[string]any) (json.RawMessage, error) {
	ctx = context.WithValue(ctx, "machineKeyID", e.id)
	ctx = context.WithValue(ctx, "machineKeyName", e.name)
	ctx = context.WithValue(ctx, "machineScopes", e.scopes)
	return e.inner.Dispatch(ctx, tool, args)
}

// NewMuxExecutor creates an executor that routes read tool calls to the given
// internal handler (an *http.ServeMux).
func NewMuxExecutor(handler http.Handler) *MuxExecutor {
	return &MuxExecutor{handler: handler}
}

// NewMuxExecutorWithWriteHandler adds the explicitly scoped MCP write surface.
// The regular handler remains GET-only; only catalog tools marked Write can
// reach writeHandler.
func NewMuxExecutorWithWriteHandler(handler http.Handler, writeHandler http.Handler) *MuxExecutor {
	return &MuxExecutor{handler: handler, writeHandler: writeHandler}
}

// Dispatch builds a GET request for the tool and executes it against the
// internal mux. Tool arguments become query parameters (or path segments for
// tools declaring PathArgs). The machine-key identity is propagated into the
// request context so read-only handlers can resolve role/user context.
func (e *MuxExecutor) Dispatch(ctx context.Context, tool ToolSpec, args map[string]any) (json.RawMessage, error) {
	if err := normalizeArgs(tool, args); err != nil {
		return nil, err
	}

	path, err := buildToolPath(tool, args)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	for _, a := range tool.Args {
		if inPath(tool.PathArgs, a.Name) || !tool.Write && a.Type == ArgObject {
			continue
		}
		if tool.Write && !inPath(tool.QueryArgs, a.Name) {
			continue
		}
		v, ok := args[a.Name]
		if !ok || v == nil {
			continue
		}
		query.Set(a.Name, stringify(v))
	}
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	method := http.MethodGet
	requestHandler := e.handler
	var requestBody []byte
	var requestContentType string
	if tool.Write {
		if e.writeHandler == nil {
			return nil, fmt.Errorf("tool %s has no write handler", tool.Name)
		}
		method = tool.Method
		if method == "" {
			method = http.MethodPost
		}
		requestHandler = e.writeHandler
		var err error
		if tool.Name == "smm_queue_create" && hasNonEmptyArgument(args["media_files"]) {
			requestBody, requestContentType, err = buildMultipartQueueBody(args)
			if err != nil {
				return nil, ToolError{Status: http.StatusBadRequest, Err: fmt.Errorf("build multipart media upload: %w", err)}
			}
		} else if payload, ok := args["payload"]; ok {
			requestBody, err = json.Marshal(payload)
			if err != nil {
				return nil, ToolError{Status: http.StatusBadRequest, Err: fmt.Errorf("encode tool payload: %w", err)}
			}
		} else {
			writeArgs := make(map[string]any, len(args))
			for key, value := range args {
				if inPath(tool.PathArgs, key) || inPath(tool.QueryArgs, key) {
					continue
				}
				if key == "media_files" {
					continue
				}
				if (key == "target_platforms" || key == "media_urls") && value != nil {
					if text, ok := value.(string); ok {
						items := make([]string, 0)
						for _, item := range strings.Split(text, ",") {
							if trimmed := strings.TrimSpace(item); trimmed != "" {
								items = append(items, trimmed)
							}
						}
						writeArgs[key] = items
						continue
					}
				}
				writeArgs[key] = value
			}
			if len(writeArgs) > 0 {
				requestBody, err = json.Marshal(writeArgs)
				if err != nil {
					return nil, ToolError{Status: http.StatusBadRequest, Err: fmt.Errorf("encode tool arguments: %w", err)}
				}
			}
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(requestBody))
	if tool.Write {
		if requestContentType == "" {
			requestContentType = "application/json"
		}
		req.Header.Set("Content-Type", requestContentType)
	}
	req = req.WithContext(withIdentity(ctx, req.Context()))

	rec := httptest.NewRecorder()
	requestHandler.ServeHTTP(rec, req)

	if rec.Code >= 400 {
		return nil, ToolError{Status: rec.Code, Err: fmt.Errorf("tool %s failed: HTTP %d: %s", tool.Name, rec.Code, rec.Body.String())}
	}
	body := rec.Body.Bytes()
	if len(body) == 0 {
		body = []byte("{}")
	}
	return sanitizeResponse(body)
}

// buildMultipartQueueBody maps the MCP queue arguments to the same multipart
// contract used by the web uploader: scalar form fields plus repeated `files`
// parts. File paths are intentionally local to the MCP process; a remote HTTP
// MCP client must first make the files available to that process.
func buildMultipartQueueBody(args map[string]any) ([]byte, string, error) {
	paths, err := mediaFilePaths(args["media_files"])
	if err != nil {
		return nil, "", err
	}
	if len(paths) == 0 {
		return nil, "", fmt.Errorf("media_files must contain at least one file")
	}
	if hasNonEmptyArgument(args["media_urls"]) {
		return nil, "", fmt.Errorf("media_files and media_urls cannot be used together")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, key := range []string{"caption", "hashtags", "post_type"} {
		if value, ok := args[key]; ok && value != nil {
			if err := writer.WriteField(key, stringify(value)); err != nil {
				return nil, "", fmt.Errorf("write %s field: %w", key, err)
			}
		}
	}
	if value, ok := args["target_platforms"]; ok && value != nil {
		platforms, err := queueTargetPlatforms(value)
		if err != nil {
			return nil, "", err
		}
		if platforms != "" {
			if err := writer.WriteField("target_platforms", platforms); err != nil {
				return nil, "", fmt.Errorf("write target_platforms field: %w", err)
			}
		}
	}

	var totalBytes int64
	for _, path := range paths {
		name, data, contentType, err := readQueueMediaFile(path)
		if err != nil {
			return nil, "", err
		}
		totalBytes += int64(len(data))
		if totalBytes > maxMCPQueueMediaTotalBytes {
			return nil, "", fmt.Errorf("media files exceed the %d MB total limit", maxMCPQueueMediaTotalBytes/(1<<20))
		}

		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
			"name":     "files",
			"filename": name,
		}))
		header.Set("Content-Type", contentType)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, "", fmt.Errorf("create file part for %q: %w", name, err)
		}
		if _, err := part.Write(data); err != nil {
			return nil, "", fmt.Errorf("write file %q: %w", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("close multipart body: %w", err)
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}

func mediaFilePaths(value any) ([]string, error) {
	var rawPaths []any
	switch typed := value.(type) {
	case []any:
		rawPaths = typed
	case []string:
		for _, path := range typed {
			rawPaths = append(rawPaths, path)
		}
	case string:
		if strings.TrimSpace(typed) != "" {
			rawPaths = []any{typed}
		}
	default:
		return nil, fmt.Errorf("media_files must be an array of local file paths")
	}
	if len(rawPaths) == 0 {
		return nil, fmt.Errorf("media_files must contain at least one file")
	}
	if len(rawPaths) > maxMCPQueueMediaFiles {
		return nil, fmt.Errorf("a maximum of %d media files is allowed", maxMCPQueueMediaFiles)
	}

	paths := make([]string, 0, len(rawPaths))
	for _, rawPath := range rawPaths {
		path, ok := rawPath.(string)
		path = strings.TrimSpace(path)
		if !ok || path == "" {
			return nil, fmt.Errorf("every media_files item must be a non-empty file path")
		}
		if !filepath.IsAbs(path) {
			return nil, fmt.Errorf("media file path %q must be absolute", path)
		}
		paths = append(paths, filepath.Clean(path))
	}
	return paths, nil
}

func readQueueMediaFile(path string) (string, []byte, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", nil, "", fmt.Errorf("open media file %q: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", nil, "", fmt.Errorf("stat media file %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return "", nil, "", fmt.Errorf("media file %q is not a regular file", path)
	}
	if info.Size() > maxMCPQueueMediaBytes {
		return "", nil, "", fmt.Errorf("media file %q exceeds the %d MB limit", path, maxMCPQueueMediaBytes/(1<<20))
	}

	data, err := io.ReadAll(io.LimitReader(file, maxMCPQueueMediaBytes+1))
	if err != nil {
		return "", nil, "", fmt.Errorf("read media file %q: %w", path, err)
	}
	if len(data) == 0 {
		return "", nil, "", fmt.Errorf("media file %q is empty", path)
	}
	if len(data) > maxMCPQueueMediaBytes {
		return "", nil, "", fmt.Errorf("media file %q exceeds the %d MB limit", path, maxMCPQueueMediaBytes/(1<<20))
	}

	name := filepath.Base(path)
	contentType := queueMediaContentType(name, data)
	if contentType == "" {
		return "", nil, "", fmt.Errorf("unsupported media type for %q; use JPG, PNG, WEBP, MP4, or MOV", name)
	}
	return name, data, contentType, nil
}

func queueMediaContentType(name string, data []byte) string {
	detected := http.DetectContentType(data)
	switch detected {
	case "image/jpeg", "image/png", "image/webp", "video/mp4", "video/quicktime":
		return detected
	}
	if detected != "application/octet-stream" {
		return ""
	}
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	default:
		return ""
	}
}

func queueTargetPlatforms(value any) (string, error) {
	var platforms []string
	switch typed := value.(type) {
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return "", nil
		}
		if strings.HasPrefix(text, "[") {
			if err := json.Unmarshal([]byte(text), &platforms); err != nil {
				return "", fmt.Errorf("target_platforms must be comma-separated or a JSON array: %w", err)
			}
		} else {
			for _, item := range strings.Split(text, ",") {
				if item = strings.TrimSpace(item); item != "" {
					platforms = append(platforms, item)
				}
			}
		}
	case []string:
		platforms = typed
	case []any:
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return "", fmt.Errorf("target_platforms items must be strings")
			}
			if text = strings.TrimSpace(text); text != "" {
				platforms = append(platforms, text)
			}
		}
	default:
		return "", fmt.Errorf("target_platforms must be a string or array of strings")
	}
	for i := range platforms {
		platforms[i] = strings.TrimSpace(platforms[i])
	}
	encoded, err := json.Marshal(platforms)
	if err != nil {
		return "", fmt.Errorf("encode target_platforms: %w", err)
	}
	return string(encoded), nil
}

func hasNonEmptyArgument(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case []string:
		return len(typed) > 0
	case []any:
		return len(typed) > 0
	default:
		return true
	}
}

// buildToolPath assembles the request path, injecting path args (in declared
// order) after the route prefix.
func buildToolPath(tool ToolSpec, args map[string]any) (string, error) {
	path := strings.TrimRight(tool.Route, "/")
	for _, name := range tool.PathArgs {
		v, ok := args[name]
		if !ok || v == nil {
			return "", fmt.Errorf("missing required path argument %q for tool %s", name, tool.Name)
		}
		path += "/" + url.PathEscape(stringify(v))
	}
	return path, nil
}

func inPath(pathArgs []string, name string) bool {
	for _, p := range pathArgs {
		if p == name {
			return true
		}
	}
	return false
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(t)
	}
}

// withIdentity carries the machine-key identity into the request context so
// handlers that read user context (role, username, userID) behave correctly.
func withIdentity(mcpCtx, reqCtx context.Context) context.Context {
	ctx := reqCtx

	if id, ok := mcpCtx.Value("machineKeyID").(int64); ok {
		ctx = context.WithValue(ctx, "userID", id)
	}
	if name, ok := mcpCtx.Value("machineKeyName").(string); ok {
		ctx = context.WithValue(ctx, "username", name)
	}
	ctx = context.WithValue(ctx, "userRole", "read")

	// Preserve machine key scopes for downstream scope checks.
	if scopes, ok := mcpCtx.Value("machineScopes").([]string); ok {
		ctx = context.WithValue(ctx, "machineScopes", scopes)
	}
	return ctx
}
