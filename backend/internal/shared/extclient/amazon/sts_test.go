package amazon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockTransport struct {
	serverURL string
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())
	reqClone.URL.Scheme = "http"
	reqClone.URL.Host = strings.TrimPrefix(m.serverURL, "http://")
	return http.DefaultTransport.RoundTrip(reqClone)
}

func TestSTSSigner_AssumeRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		xmlResponse := `
		<AssumeRoleResponse>
		  <AssumeRoleResult>
			<Credentials>
			  <AccessKeyId>ASIA_TEST</AccessKeyId>
			  <SecretAccessKey>SECRET_TEST</SecretAccessKey>
			  <SessionToken>SESSION_TEST</SessionToken>
			  <Expiration>2032-05-09T16:00:00Z</Expiration>
			</Credentials>
		  </AssumeRoleResult>
		</AssumeRoleResponse>`
		w.Write([]byte(strings.TrimSpace(xmlResponse)))
	}))
	defer server.Close()

	signer := NewSTSSigner("AKIA", "SECRET", "us-east-1", "arn:aws:iam::123:role/TestRole")
	signer.httpClient.Transport = &mockTransport{
		serverURL: server.URL,
	}

	access, secret, session, err := signer.AssumeRole()
	if err != nil {
		t.Fatalf("AssumeRole failed: %v", err)
	}
	if access != "ASIA_TEST" || secret != "SECRET_TEST" || session != "SESSION_TEST" {
		t.Errorf("Unexpected credentials returned")
	}

	// Test no roleARN
	signerNoRole := NewSTSSigner("AKIA", "SECRET", "us-east-1", "")
	a, s, sess, err := signerNoRole.AssumeRole()
	if err != nil || a != "AKIA" || s != "SECRET" || sess != "" {
		t.Errorf("Unexpected result for empty role ARN")
	}
}
