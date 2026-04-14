package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mi-tech/internal/config"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

type MetaClient struct {
	settings   *config.SettingsProvider
	apiVersion string
}

func NewMetaClient(settings *config.SettingsProvider) *MetaClient {
	return &MetaClient{
		settings:   settings,
		apiVersion: "v22.0",
	}
}

type TemplateRequest struct {
	Name       string                   `json:"name"`
	Language   string                   `json:"language"`
	Category   string                   `json:"category"`
	Components []map[string]interface{} `json:"components"`
}

func (c *MetaClient) CreateTemplate(req TemplateRequest) (string, error) {
	wabaID := c.settings.GetWhatsAppWABAID()
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates", c.apiVersion, wabaID)
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

	httpReq.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())
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
	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())

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
	reqUpload.Header.Set("Authorization", "OAuth "+c.settings.GetWhatsAppAccessToken())
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
	phoneNumberID := c.settings.GetWhatsAppPhoneNumberID()
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/media", c.apiVersion, phoneNumberID)

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

	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())
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

	httpReq.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())
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
	wabaID := c.settings.GetWhatsAppWABAID()
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates?name=%s", c.apiVersion, wabaID, templateName)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())

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
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Category   string                   `json:"category"`
	Language   string                   `json:"language"`
	Status     string                   `json:"status"`
	Components []map[string]interface{} `json:"components"`
}

func (c *MetaClient) GetRemoteTemplateByName(templateName string) (*RemoteTemplate, error) {
	wabaID := c.settings.GetWhatsAppWABAID()
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates?name=%s", c.apiVersion, wabaID, url.QueryEscape(templateName))

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())

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

func (c *MetaClient) GetAllRemoteTemplates() ([]RemoteTemplate, error) {
	wabaID := c.settings.GetWhatsAppWABAID()
	var allTemplates []RemoteTemplate
	
	urlStr := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates?limit=100", c.apiVersion, wabaID)

	for urlStr != "" {
		httpReq, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return nil, err
		}

		httpReq.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return nil, err
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("meta api error: status %d, body: %s", resp.StatusCode, string(respBody))
		}

		var result struct {
			Data   []RemoteTemplate `json:"data"`
			Paging struct {
				Next     string `json:"next"`
				Previous string `json:"previous"`
			} `json:"paging"`
		}
		
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, err
		}

		allTemplates = append(allTemplates, result.Data...)
		
		urlStr = result.Paging.Next
	}

	return allTemplates, nil
}

func (c *MetaClient) SendTemplateMessage(phoneNumber, templateName, languageCode string, components []interface{}) (string, error) {
	phoneNumberID := c.settings.GetWhatsAppPhoneNumberID()
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.apiVersion, phoneNumberID)

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

	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())
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

	wabaID := c.settings.GetWhatsAppWABAID()
	apiURL := fmt.Sprintf("https://graph.facebook.com/%s/%s?fields=template_analytics.start(%s).end(%s)",
		c.apiVersion, wabaID, url.QueryEscape(startDate), url.QueryEscape(endDate))

	log.Printf("Fetching Meta Analytics from: %s", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())

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
				TemplateID   string `json:"template_id"`
				TemplateName string `json:"template_name"`
				Sent         int    `json:"sent"`
				Delivered    int    `json:"delivered"`
				Read         int    `json:"read"`
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

func (c *MetaClient) SendMediaMessage(phoneNumber, mediaID, mediaType, caption string) (string, error) {
	phoneNumberID := c.settings.GetWhatsAppPhoneNumberID()
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.apiVersion, phoneNumberID)

	mediaPayload := map[string]string{
		"id": mediaID,
	}
	if caption != "" && (mediaType == "image" || mediaType == "video" || mediaType == "document") {
		mediaPayload["caption"] = caption
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                phoneNumber,
		"type":              mediaType,
		mediaType:           mediaPayload,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Meta API Response (SendMedia): %s", string(respBody))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("meta api error: status %d, body: %s", resp.StatusCode, string(respBody))
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

func (c *MetaClient) SendTextMessage(phoneNumber, text string) (string, error) {
	phoneNumberID := c.settings.GetWhatsAppPhoneNumberID()
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", c.apiVersion, phoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                phoneNumber,
		"type":              "text",
		"text": map[string]string{
			"body": text,
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

	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Meta API Response (SendText): %s", string(respBody))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("meta api error: status %d, body: %s", resp.StatusCode, string(respBody))
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

func (c *MetaClient) GetMediaURL(mediaID string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s", c.apiVersion, mediaID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get media URL: %s", string(respBody))
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.URL, nil
}

func (c *MetaClient) DownloadMedia(downloadURL string) ([]byte, string, error) {
	req, _ := http.NewRequest("GET", downloadURL, nil)
	// Some media URLs already contain tokens or require the Authorization header
	req.Header.Set("Authorization", "Bearer "+c.settings.GetWhatsAppAccessToken())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("failed to download media binary: %s", string(respBody))
	}

	data, err := io.ReadAll(resp.Body)
	return data, resp.Header.Get("Content-Type"), err
}
