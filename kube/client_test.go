package kube

import (
	"testing"
)

func TestClientGetDeployment(t *testing.T) {
	client, _ := NewFake()
	client.CreateDeployment(&Deployment{Name: "myapp", Namespace: "default"})

	d, _ := client.GetDeployment("default", "myapp")

	if d.Name != "myapp" {
		t.Errorf("GetDeployment got incorrect deployment: %v", d)
	}
}

func TestClientUpdateDeploymentImage(t *testing.T) {
	client, _ := NewFake()
	client.CreateDeployment(
		&Deployment{
			Name:      "myapp",
			Namespace: "default",
			Containers: []*Container{
				{Name: "app", Image: "nginx:latest"},
			},
		},
	)

	client.UpdateDeploymentImage("default", "myapp", "nginx:1.21-alpine")

	d, _ := client.GetDeployment("default", "myapp")
	if d.Containers[0].Image != "nginx:1.21-alpine" {
		t.Errorf("UpdateDeploymentImage did not update image. Was: %s", d.Containers[0].Image)
	}
}
