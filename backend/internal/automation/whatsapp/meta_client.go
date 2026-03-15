package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

type MetaClient struct {
	accessToken   string
	phoneNumberID string
	wabaID        string
	apiVersion    string
}

func NewMetaClient(accessToken, phoneNumberID, wabaID string) *MetaClient {
	return &MetaClient{
		accessToken:   accessToken,
		phoneNumberID: phoneNumberID,
		wabaID:        wabaID,
		apiVersion:    "v22.0",
	}
}

type TemplateRequest struct {
	Name       string                   `json:"name"`
	Language   string                   `json:"language"`
	Category   string                   `json:"category"`
	Components []map[string]interface{} `json:"components"`
}

func (c *MetaClient) CreateTemplate(req TemplateRequest) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates", c.apiVersion, c.wabaID)
	log.Printf("Creating Meta Template at URL: %s", url)

	// Use a custom encoder to prevent escaping of '&' and other characters in URLs
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(req); err != nil {
		return "", err
	}
	body := buf.Bytes()
	log.Printf("Meta API Request (CreateTemplate) Payload: %s", string(body))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Meta API Response (CreateTemplate): %s", string(respBody))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("meta api error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (c *MetaClient) UploadMediaFromURL(appID string, fileURL string) (string, error) {
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("sample download failed: status %d from %s", resp.StatusCode, fileURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read media body: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	return c.UploadMediaFromBytes(appID, body, mimeType)
}

func (c *MetaClient) UploadMediaFromBytes(appID string, body []byte, mimeType string) (string, error) {
	fileLength := len(body)

	// Guess mime type, default to PDF for documents if uncertain
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = http.DetectContentType(body)
		if mimeType == "application/octet-stream" {
			mimeType = "application/pdf"
		}
	}

	// Strip parameters (like ; qs=0.001 or ; charset=utf-8) for Meta compatibility
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}

	// 2. Create upload session
	escapedMime := url.QueryEscape(mimeType)
	sessionURL := fmt.Sprintf("https://graph.facebook.com/%s/%s/uploads?file_length=%d&file_type=%s", c.apiVersion, appID, fileLength, escapedMime)
	
	log.Printf("Meta Automation: Creating upload session. Length: %d, Type: %s, URL: %s", fileLength, mimeType, sessionURL)
	
	req, _ := http.NewRequest("POST", sessionURL, nil)
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	sessionResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create upload session: %w", err)
	}
	defer sessionResp.Body.Close()

	sessionBody, _ := io.ReadAll(sessionResp.Body)
	if sessionResp.StatusCode >= 400 {
		return "", fmt.Errorf("create session failed: status %d, body: %s", sessionResp.StatusCode, string(sessionBody))
	}

	var sessionData struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(sessionBody, &sessionData); err != nil || sessionData.ID == "" {
		return "", fmt.Errorf("invalid session response: %s", string(sessionBody))
	}

	// 3. Upload file to session
	uploadURL := fmt.Sprintf("https://graph.facebook.com/%s/%s", c.apiVersion, sessionData.ID)
	reqUpload, _ := http.NewRequest("POST", uploadURL, bytes.NewBuffer(body))
	reqUpload.Header.Set("Authorization", "OAuth "+c.accessToken)
	reqUpload.Header.Set("file_offset", "0")

	uploadResp, err := http.DefaultClient.Do(reqUpload)
	if err != nil {
		return "", fmt.Errorf("failed to upload media: %w", err)
	}
	defer uploadResp.Body.Close()

	uploadResBody, _ := io.ReadAll(uploadResp.Body)
	if uploadResp.StatusCode >= 400 {
		return "", fmt.Errorf("media upload failed: status %d, body: %s", uploadResp.StatusCode, string(uploadResBody))
	}

	var uploadData struct {
		Handle string `json:"h"`
	}
	if err := json.Unmarshal(uploadResBody, &uploadData); err != nil || uploadData.Handle == "" {
		return "", fmt.Errorf("invalid upload response: %s", string(uploadResBody))
	}

	return uploadData.Handle, nil
}

func (c *MetaClient) UploadWhatsAppMedia(body []byte, filename, mimeType string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/media", c.apiVersion, c.phoneNumberID)

	// Create a multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	
	// Add messaging_product
	_ = w.WriteField("messaging_product", "whatsapp")
	
	// Add the file with explicit Content-Type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", filename))
	h.Set("Content-Type", mimeType)
	part, err := w.CreatePart(h)
	if err != nil {
		return "", err
	}
	_, _ = io.Copy(part, bytes.NewReader(body))
	
	// Add type
	_ = w.WriteField("type", mimeType)
	
	_ = w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("whatsapp media upload failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.ID, nil
}

func (c *MetaClient) UpdateTemplate(metaTemplateID string, components []map[string]interface{}) error {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s", c.apiVersion, metaTemplateID)

	payload := map[string]interface{}{
		"components": components,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Meta API Error (UpdateTemplate): status %d, body: %s", resp.StatusCode, string(respBody))
		return fmt.Errorf("meta api error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *MetaClient) DeleteTemplate(templateName string) error {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates?name=%s", c.apiVersion, c.wabaID, templateName)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Meta API Error (DeleteTemplate): %s", string(respBody))
		return fmt.Errorf("meta api error: status %d", resp.StatusCode)
	}

	return nil
}

type RemoteTemplate struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *MetaClient) GetRemoteTemplateByName(templateName string) (*RemoteTemplate, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates?name=%s", c.apiVersion, c.wabaID, url.QueryEscape(templateName))

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("meta api error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []RemoteTemplate `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, nil // Not found
	}

	return &result.Data[0], nil
}

func (c *MetaClient) SendTemplateMessage(phoneNumber, templateName, languageCode string, components []interface{}) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.apiVersion, c.phoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                phoneNumber,
		"type":              "template",
		"template": map[string]interface{}{
			"name": templateName,
			"language": map[string]string{
				"code": languageCode,
			},
			"components": components,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Meta API Response (SendTemplate): %s", string(respBody))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("meta api error: status %d", resp.StatusCode)
	}

	var result struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Messages) > 0 {
		return result.Messages[0].ID, nil
	}

	return "", nil
}

type MetaTemplateAnalytics struct {
	Data []struct {
		TemplateID string `json:"template_id"`
		Name       string `json:"name"`
		Sent       int    `json:"sent"`
		Delivered  int    `json:"delivered"`
		Read       int    `json:"read"`
	} `json:"data"`
}

func (c *MetaClient) GetTemplateAnalytics(startDate, endDate string) (map[string]AutomationTemplate, error) {
	// Format: GET /WABA_ID?fields=template_analytics.start(YYYY-MM-DD).end(YYYY-MM-DD).granularity(DAILY)
	// Note: Meta API requires specific date formats.
	
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s?fields=template_analytics.start(%s).end(%s)", 
		c.apiVersion, c.wabaID, url.QueryEscape(startDate), url.QueryEscape(endDate))
	
	log.Printf("Fetching Meta Analytics from: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read analytics response: %w", err)
	}

	if resp.StatusCode >= 400 {
		log.Printf("Meta Analytics API Error: %s", string(respBody))
		return nil, fmt.Errorf("meta api error: status %d", resp.StatusCode)
	}

	var result struct {
		TemplateAnalytics struct {
			Data []struct {
				TemplateID string `json:"template_id"`
				TemplateName string `json:"template_name"`
				Sent       int    `json:"sent"`
				Delivered  int    `json:"delivered"`
				Read       int    `json:"read"`
			} `json:"data"`
		} `json:"template_analytics"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	analyticsMap := make(map[string]AutomationTemplate)
	for _, d := range result.TemplateAnalytics.Data {
		analyticsMap[d.TemplateName] = AutomationTemplate{
			TemplateName:   d.TemplateName,
			SentCount:      d.Sent,
			DeliveredCount: d.Delivered,
			ReadCount:      d.Read,
		}
	}

	return analyticsMap, nil
}
