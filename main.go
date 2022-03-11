package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

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

	r := buildRouter(conf, logger, kube)

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

func handlePushArtifact(c *config.Config, logger *zap.Logger, api kubernetes.Interface, w *HarborWebhookEvent) error {
	logger.Info("Received Harbor webhook", zap.String("image_name", w.Repository.FullName))
	for _, m := range c.Mappings {
		if w.Repository.FullName == m.ImageName {
			client := api.AppsV1().Deployments(m.Namespace)
			deployment, err := client.Get(context.TODO(), m.DeploymentName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			// TODO: support specifying which container to update; support multiple resources(?)
			deployment.Spec.Template.Spec.Containers[0].Image = w.Resources[0].ResourceURL
			_, err = client.Update(context.TODO(), deployment, metav1.UpdateOptions{})
			logger.Info("Updated deployment",
				zap.String("image_name", m.ImageName), zap.String("deployment", m.DeploymentName))
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

func buildRouter(conf *config.Config, logger *zap.Logger, kube kubernetes.Interface) *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), requestLogger(logger), auth(conf))

	r.POST("/webhooks/harbor", func(c *gin.Context) {
		var webhook HarborWebhook
		err := c.BindJSON(&webhook)
		if err != nil {
			logger.Info("JSON unmarshal failure", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"ok": false})
			return
		}
		if webhook.EventType == "PUSH_ARTIFACT" {
			err := handlePushArtifact(conf, logger, kube, &webhook.EventData)
			if err != nil {
				logger.Info("Error while updating deployment", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"ok": false})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}
