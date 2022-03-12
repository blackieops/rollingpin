package kube

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Deployment struct {
	Namespace  string
	Name       string
	Containers []*Container
}

type Container struct {
	Name  string
	Image string
}

func (d *Deployment) FromKubernetes(kd *appsv1.Deployment) {
	var containers []*Container
	for _, c := range kd.Spec.Template.Spec.Containers {
		containers = append(containers, &Container{Name: c.Name, Image: c.Image})
	}
	d.Namespace = kd.Namespace
	d.Name = kd.Name
	d.Containers = containers
}

func (d *Deployment) ToKubernetes() *appsv1.Deployment {
	var containers []v1.Container
	for _, c := range d.Containers {
		containers = append(containers, v1.Container{Name: c.Name, Image: c.Image})
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: d.Name, Namespace: d.Namespace},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: containers,
				},
			},
		},
	}
}
