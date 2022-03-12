package main

import (
	"flag"
	"time"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.b8s.dev/rollingpin/kube"
	"go.b8s.dev/rollingpin/providers/harbor"
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

func buildRouter(conf *config.Config, logger *zap.Logger, client kube.IClient) *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(gin.Recovery(), requestLogger(logger))

	harborRouter := &harbor.Router{Config: conf, Logger: logger, Client: client}
	harborRouter.Mount(r.Group("/webhooks/harbor"))

	return r
}
