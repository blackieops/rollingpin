package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.b8s.dev/rollingpin/kube"
	"go.uber.org/zap"
)

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
	// Set up test HTTP request
	payload := `{
		"type": "PUSH_ARTIFACT",
		"occur_at": 1586922308,
		"operator": "admin",
		"event_data": {
			"resources": [{
				"digest": "sha256:8a9e9863dbb6e10edb5adfe917c00da84e1700fa76e7ed02476aa6e6fb8ee0d8",
				"tag": "latest",
				"resource_url": "cr.b8s.dev/library/debian:v2"
			}],
			"repository": {
				"date_created": 1586922308,
				"name": "debian",
				"namespace": "library",
				"repo_full_name": "library/debian",
				"repo_type": "private"
			}
		}
	}`
	req, _ := http.NewRequest("POST", "/webhooks/harbor", bytes.NewBufferString(payload))
	req.Header.Add("authorization", "Bearer abc1234")
	resp := httptest.NewRecorder()

	// Set up fake kubernetes api client
	fakeClient, _ := kube.NewFake()
	fakeClient.CreateDeployment(
		&kube.Deployment{
			Namespace: "default",
			Name:      "test-deployment",
			Containers: []*kube.Container{
				{Name: "app", Image: "cr.b8s.dev/library/debian:v1"},
			},
		},
	)

	// Set up our own app's stuff
	conf := &config.Config{
		AuthToken: "abc1234",
		Mappings: []config.ImageMapping{
			{
				Namespace:      "default",
				DeploymentName: "test-deployment",
				ImageName:      "library/debian",
			},
		},
	}
	log, _ := zap.NewProduction()

	// Execute request
	r := buildRouter(conf, log, fakeClient)
	r.ServeHTTP(resp, req)

	// Assertions
	if resp.Code != 200 {
		t.Errorf("Expected 200 response got: %d", resp.Code)
	}
	if resp.Body.String() != `{"ok":true}` {
		t.Errorf("Expected OK response got: %s", resp.Body.String())
	}

	newDeploy, _ := fakeClient.GetDeployment("default", "test-deployment")
	newImageName := newDeploy.Containers[0].Image
	if newImageName != "cr.b8s.dev/library/debian:v2" {
		t.Errorf("Expected deployment to be updated but was not! Image was: %s", newImageName)
	}
}

func buildTestConn(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	conn, _ := gin.CreateTestContext(w)
	conn.Request = req
	return conn, w
}
