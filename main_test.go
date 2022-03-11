package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	var webhook HarborWebhook
	err := json.Unmarshal([]byte(payload), &webhook)
	if err != nil {
		t.Errorf("Failed to unmarshal HarborWebhook: %v", err)
		return
	}
	if webhook.EventType != "PUSH_ARTIFACT" {
		t.Errorf("HarborWebhook unmarshalled incorrect type: %v", webhook.EventType)
	}
	if webhook.EventData.Repository.FullName != "test-webhook/debian" {
		t.Errorf("HarborWebhookEvent.Repository had incorrect FullName: %v", webhook.EventData.Repository.FullName)
	}
	if webhook.EventData.Resources[0].ResourceURL != "hub.harbor.com/test-webhook/debian:latest" {
		t.Errorf(
			"HarborWebhookEvent.Resources[0] had incorrect ResourceURL: %v",
			webhook.EventData.Resources[0].ResourceURL,
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api := fake.NewSimpleClientset()

	// Set up initial kubernetes deployment
	initialDeploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-deployment"},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "app", Image: "cr.b8s.dev/library/debian:v1"},
					},
				},
			},
		},
	}
	api.AppsV1().Deployments("default").Create(ctx, initialDeploy, metav1.CreateOptions{})

	// Set up our own app's stuff finally
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
	r := buildRouter(conf, log, api)
	r.ServeHTTP(resp, req)

	// Assertions
	if resp.Code != 200 {
		t.Errorf("Expected 200 response got: %d", resp.Code)
	}
	if resp.Body.String() != `{"ok":true}` {
		t.Errorf("Expected OK response got: %s", resp.Body.String())
	}

	newDeploy, _ := api.AppsV1().Deployments("default").Get(ctx, "test-deployment", metav1.GetOptions{})
	newImageName := newDeploy.Spec.Template.Spec.Containers[0].Image
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
