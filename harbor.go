package main

type HarborWebhook struct {
	EventType string             `json:"type"`
	OccurAt   int                `json:"occur_at"`
	Operator  string             `json:"operator"`
	EventData HarborWebhookEvent `json:"event_data"`
}

type HarborWebhookEvent struct {
	Resources  []HarborWebhookResource `json:"resources"`
	Repository HarborWebhookRepository `json:"repository"`
}

type HarborWebhookRepository struct {
	DateCreated int    `json:"date_created"`
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	FullName    string `json:"repo_full_name"`
	Type        string `json:"repo_type"`
}

type HarborWebhookResource struct {
	Digest      string `json:"digest"`
	Tag         string `json:"tag"`
	ResourceURL string `json:"resource_url"`
}
