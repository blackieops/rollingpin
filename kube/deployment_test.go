package kube

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestDeploymentFromKubernetes(t *testing.T) {
	source := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "myapp", Namespace: "default"},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: "app", Image: "nginx:latest"},
						{Name: "proxy", Image: "envoy:latest"},
					},
				},
			},
		},
	}

	d := &Deployment{}
	d.FromKubernetes(source)

	if d.Name != "myapp" {
		t.Errorf("FromKubernetes set incorrect name: %s", d.Name)
	}
	if d.Namespace != "default" {
		t.Errorf("FromKubernetes set incorrect namespace: %s", d.Namespace)
	}
	if d.Containers[0].Name != "app" {
		t.Errorf("FromKubernetes set incorrect container name: %s", d.Containers[0].Name)
	}
	if d.Containers[0].Image != "nginx:latest" {
		t.Errorf("FromKubernetes set incorrect container image: %s", d.Containers[0].Image)
	}
	if d.Containers[1].Name != "proxy" {
		t.Errorf("FromKubernetes set incorrect container name: %s", d.Containers[1].Name)
	}
	if d.Containers[1].Image != "envoy:latest" {
		t.Errorf("FromKubernetes set incorrect container image: %s", d.Containers[1].Image)
	}
}

func TestDeploymentToKubernetes(t *testing.T) {
	source := &Deployment{
		Namespace: "default",
		Name:      "myapp",
		Containers: []*Container{
			{Name: "app", Image: "nginx:latest"},
			{Name: "proxy", Image: "envoy:latest"},
		},
	}

	d := source.ToKubernetes()

	if d.Name != "myapp" {
		t.Errorf("ToKubernetes set incorrect name: %s", d.Name)
	}
	if d.Namespace != "default" {
		t.Errorf("ToKubernetes set incorrect namespace: %s", d.Namespace)
	}
	if d.Spec.Template.Spec.Containers[0].Name != "app" {
		t.Errorf(
			"ToKubernetes set incorrect container name: %s",
			d.Spec.Template.Spec.Containers[0].Name,
		)
	}
	if d.Spec.Template.Spec.Containers[0].Image != "nginx:latest" {
		t.Errorf(
			"ToKubernetes set incorrect container image: %s",
			d.Spec.Template.Spec.Containers[0].Image,
		)
	}
	if d.Spec.Template.Spec.Containers[1].Name != "proxy" {
		t.Errorf(
			"ToKubernetes set incorrect container name: %s",
			d.Spec.Template.Spec.Containers[1].Name,
		)
	}
	if d.Spec.Template.Spec.Containers[1].Image != "envoy:latest" {
		t.Errorf(
			"ToKubernetes set incorrect container image: %s",
			d.Spec.Template.Spec.Containers[1].Image,
		)
	}
}
