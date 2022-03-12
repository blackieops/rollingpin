package kube

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

type IClient interface {
	GetDeployment(string, string) (*Deployment, error)
	UpdateDeploymentImage(string, string, string) error
	CreateDeployment(*Deployment) error
}

type Client struct {
	clientset kubernetes.Interface
}

func New() (*Client, error) {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kube, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, err
	}
	return &Client{clientset: kube}, nil
}

func NewFake() (*Client, error) {
	return &Client{clientset: fake.NewSimpleClientset()}, nil
}

func (c *Client) GetDeployment(ns string, name string) (*Deployment, error) {
	kdeploy, err := c.clientset.AppsV1().Deployments(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	deployment := &Deployment{}
	deployment.FromKubernetes(kdeploy)
	return deployment, nil
}

func (c *Client) UpdateDeploymentImage(ns string, name string, image string) error {
	client := c.clientset.AppsV1().Deployments(ns)
	deployment, err := client.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// TODO: support specifying which container to update
	deployment.Spec.Template.Spec.Containers[0].Image = image
	_, err = client.Update(context.TODO(), deployment, metav1.UpdateOptions{})
	return nil
}

// XXX: this is purely for testing, feels a bit weird to have it as part of the
// official interface. Our deployment objects are too minimal to really create
// a legitimate deployment from scratch.
func (c *Client) CreateDeployment(d *Deployment) error {
	deploy := d.ToKubernetes()
	_, err := c.clientset.AppsV1().Deployments(d.Namespace).Create(context.TODO(), deploy, metav1.CreateOptions{})
	return err
}
