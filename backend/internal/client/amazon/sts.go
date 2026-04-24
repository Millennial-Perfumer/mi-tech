package amazon

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// AssumeRoleResponse represents the STS AssumeRole response.
type AssumeRoleResponse struct {
	AssumeRoleResult struct {
		Credentials struct {
			AccessKeyId     string `xml:"AccessKeyId"`
			SecretAccessKey string `xml:"SecretAccessKey"`
			SessionToken    string `xml:"SessionToken"`
			Expiration      string `xml:"Expiration"`
		} `xml:"Credentials"`
	} `xml:"AssumeRoleResult"`
}

// STSSigner handles AWS STS AssumeRole to get temporary credentials.
type STSSigner struct {
	accessKey string
	secretKey string
	region    string
	roleARN   string
}

func NewSTSSigner(accessKey, secretKey, region, roleARN string) *STSSigner {
	return &STSSigner{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
		roleARN:   roleARN,
	}
}

// AssumeRole calls AWS STS to get temporary credentials for the SP-API role.
func (s *STSSigner) AssumeRole() (accessKey, secretKey, sessionToken string, err error) {
	if s.roleARN == "" {
		return s.accessKey, s.secretKey, "", nil
	}

	endpoint := fmt.Sprintf("https://sts.%s.amazonaws.com", s.region)
	data := url.Values{}
	data.Set("Action", "AssumeRole")
	data.Set("Version", "2011-06-15")
	data.Set("RoleArn", s.roleARN)
	data.Set("RoleSessionName", "MI_App_SPAPI_Session")
	data.Set("DurationSeconds", "3600")

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// STS requires SigV4 signing too
	err = SignV4(req, []byte(data.Encode()), s.accessKey, s.secretKey, s.region, "sts", time.Now())
	if err != nil {
		return "", "", "", fmt.Errorf("failed to sign STS request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("STS AssumeRole error (%d): %s", resp.StatusCode, string(body))
	}

	var assumeResp AssumeRoleResponse
	if err := xml.Unmarshal(body, &assumeResp); err != nil {
		return "", "", "", err
	}

	cred := assumeResp.AssumeRoleResult.Credentials
	return cred.AccessKeyId, cred.SecretAccessKey, cred.SessionToken, nil
}
