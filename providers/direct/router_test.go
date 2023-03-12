package direct

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
)

func TestUnmarshalWebhook(t *testing.T) {
	payload := `{
		"image_url": "cr.example.com/test-webhook/debian:latest",
		"repository_name": "test-webhook/debian"
	}`
	var webhook DirectWebhook
	err := json.Unmarshal([]byte(payload), &webhook)
	if err != nil {
		t.Errorf("Failed to unmarshal DirectWebhook: %v", err)
		return
	}
	if webhook.RepositoryName != "test-webhook/debian" {
		t.Errorf("DirectWebhook had wrong RepositoryName: %v", webhook.RepositoryName)
	}
	if webhook.ImageURL != "cr.example.com/test-webhook/debian:latest" {
		t.Errorf(
			"DirectWebhook had incorrect ImageURL: %v",
			webhook.ImageURL,
		)
	}
}

func TestAuthSuccess(t *testing.T) {
	conf := &config.Config{
		AuthToken: "test1234",
	}
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(""))
	req.Header.Add("authorization", "Bearer test1234")
	ctx, resp := buildTestConn(req)
	r := &Router{Config: conf}
	r.auth()(ctx)
	if resp.Code != 200 {
		t.Errorf("Request should have been 200 but was %d!", resp.Code)
	}
}

func TestAuthMismatch(t *testing.T) {
	conf := &config.Config{
		AuthToken: "asdfljasfdadsfjasdf",
	}
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(""))
	req.Header.Add("authorization", "Bearer test1234")
	ctx, resp := buildTestConn(req)
	r := &Router{Config: conf}
	r.auth()(ctx)
	if resp.Code != 401 {
		t.Errorf("Request should have been 401 but was %d!", resp.Code)
	}
}

func buildTestConn(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	conn, _ := gin.CreateTestContext(w)
	conn.Request = req
	return conn, w
}
