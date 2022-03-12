package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.b8s.dev/rollingpin/kube"
	"go.uber.org/zap"
)

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
	kube, err := kube.New()
	if err != nil {
		panic(err)
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

func handlePushArtifact(c *config.Config, logger *zap.Logger, api kube.IClient, w *HarborWebhookEvent) error {
	logger.Info("Received Harbor webhook", zap.String("image_name", w.Repository.FullName))
	for _, m := range c.Mappings {
		if w.Repository.FullName == m.ImageName {
			err := api.UpdateDeploymentImage(m.Namespace, m.DeploymentName, w.Resources[0].ResourceURL)
			logger.Info("Updated deployment",
				zap.String("image_name", m.ImageName),
				zap.String("deployment", m.DeploymentName))
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

func buildRouter(conf *config.Config, logger *zap.Logger, client kube.IClient) *gin.Engine {
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
			err := handlePushArtifact(conf, logger, client, &webhook.EventData)
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
