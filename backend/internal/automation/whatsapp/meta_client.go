package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
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
		return "", fmt.Errorf("meta api error: status %d", resp.StatusCode)
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
		log.Printf("Meta API Error (UpdateTemplate): %s", string(respBody))
		return fmt.Errorf("meta api error: status %d", resp.StatusCode)
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

func (c *MetaClient) GetTemplateStatus(templateName string) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/message_templates?name=%s", c.apiVersion, c.wabaID, templateName)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("meta api error: status %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			Status string `json:"status"`
			Name   string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	return result.Data[0].Status, nil
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
	if resp.StatusCode >= 400 {
		log.Printf("Meta API Error (SendTemplate): %s", string(respBody))
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
