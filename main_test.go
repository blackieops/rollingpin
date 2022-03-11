package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUnmarshalHarborWebhook(t *testing.T) {
	payload := `{
		"type": "PUSH_ARTIFACT",
		"occur_at": 1586922308,
		"operator": "admin",
		"event_data": {
			"resources": [{
				"digest": "sha256:8a9e9863dbb6e10edb5adfe917c00da84e1700fa76e7ed02476aa6e6fb8ee0d8",
				"tag": "latest",
				"resource_url": "hub.harbor.com/test-webhook/debian:latest"
			}],
			"repository": {
				"date_created": 1586922308,
				"name": "debian",
				"namespace": "test-webhook",
				"repo_full_name": "test-webhook/debian",
				"repo_type": "private"
			}
		}
	}`
	var webhook *HarborWebhook
	err := json.Unmarshal([]byte(payload), &webhook)
	if err != nil {
		t.Errorf("Failed to unmarshal HarborWebhook: %v", err)
		return
	}
	if webhook.EventType != "PUSH_ARTIFACT" {
		t.Errorf("HarborWebhook unmarshalled incorrect type: %v", webhook.EventType)
	}
}

func TestAuthSuccess(t *testing.T) {
	conf := &config.Config{
		AuthToken: "test1234",
	}
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(""))
	req.Header.Add("authorization", "Bearer test1234")
	ctx, resp := buildTestConn(req)
	auth(conf)(ctx)
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
	auth(conf)(ctx)
	if resp.Code != 401 {
		t.Errorf("Request should have been 401 but was %d!", resp.Code)
	}
}

func TestHarborWebhookPushArtifact(t *testing.T) {
	conf := &config.Config{AuthToken: "abc1234"}
	api := fake.NewSimpleClientset()
	log, _ := zap.NewProduction()
	payload := `{
		"type": "PUSH_ARTIFACT",
		"occur_at": 1586922308,
		"operator": "admin",
		"event_data": {
			"resources": [{
				"digest": "sha256:8a9e9863dbb6e10edb5adfe917c00da84e1700fa76e7ed02476aa6e6fb8ee0d8",
				"tag": "latest",
				"resource_url": "hub.harbor.com/test-webhook/debian:latest"
			}],
			"repository": {
				"date_created": 1586922308,
				"name": "debian",
				"namespace": "test-webhook",
				"repo_full_name": "test-webhook/debian",
				"repo_type": "private"
			}
		}
	}`
	req, _ := http.NewRequest("POST", "/webhooks/harbor", bytes.NewBufferString(payload))
	req.Header.Add("authorization", "Bearer abc1234")
	resp := httptest.NewRecorder()
	r := buildRouter(conf, log, api)
	r.ServeHTTP(resp, req)
	if resp.Code != 200 {
		t.Errorf("Expected 200 response got: %d", resp.Code)
	}
	if resp.Body.String() != `{"ok":true}` {
		t.Errorf("Expected OK response got: %s", resp.Body.String())
	}
}

func buildTestConn(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	conn, _ := gin.CreateTestContext(w)
	conn.Request = req
	return conn, w
}
