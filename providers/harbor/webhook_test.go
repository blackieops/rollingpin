package harbor

import (
	"encoding/json"
	"testing"
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
