package main

import (
	"fmt"
	"context"
	"flag"
	"net/http"
	"time"

	"go.b8s.dev/rollingpin/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type HarborWebhook struct {
	EventType string               `json:"event_type"`
	Events    []HarborWebhookEvent `json:"events"`
}

type HarborWebhookEvent struct {
	Project     string `json:"project"`
	RepoName    string `json:"repo_name"`
	Tag         string `json:"tag"`
	FullName    string `json:"full_name"`
	TriggerTime int    `json:"trigger_time"`
	ImageID     string `json:"image_id"`
	ProjectType string `json:"project_type"`
}

var configPath = flag.String("config", "config.yaml", "Path to the config file.")

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	conf, err := config.LoadConfig(*configPath)
	if err != nil {
		panic(err)
	}
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	kube, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		panic(err.Error())
	}

	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), requestLogger(logger))

	r.POST("/webhooks/harbor", auth(conf), func(c *gin.Context) {
		var webhook *HarborWebhook
		err := c.BindJSON(webhook)
		if err != nil {
			logger.Info("JSON unmarshal failure", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"ok": false})
			return
		}
		if webhook.EventType == "pushImage" {
			for _, e := range webhook.Events {
				go handlePushImage(conf, logger, kube, &e)
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	logger.Info("Server started on :8080.")
	r.Run(":8080")
}

func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestStart := time.Now()
		c.Next()

		logger.Info("Request Processed",
			zap.String("method", c.Request.Method),
			zap.Int("status_code", c.Writer.Status()),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("latency", time.Since(requestStart)),
		)
	}
}

func handlePushImage(c *config.Config, logger *zap.Logger, api *kubernetes.Clientset, w *HarborWebhookEvent) error {
	for _, m := range c.Mappings {
		if w.FullName == m.ImageName {
			logger.Info("Webhook matched configured deployment",
				zap.String("image_name", m.ImageName), zap.String("deployment", m.DeploymentName))
			client := api.AppsV1().Deployments(m.Namespace)
			deployment, err := client.Get(context.TODO(), m.DeploymentName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			// TODO: support specifying which container to update
			deployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", w.FullName, w.Tag)
			_, err = client.Update(context.TODO(), deployment, metav1.UpdateOptions{})
			return err
		}
	}
	return nil
}

func auth(conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerValue := c.Request.Header.Get("Authorization")
		if headerValue == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		token := headerValue[len("Bearer "):]
		if token != conf.AuthToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
