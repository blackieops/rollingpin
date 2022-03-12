package harbor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.b8s.dev/rollingpin/kube"
	"go.uber.org/zap"
)

type Router struct {
	Config *config.Config
	Logger *zap.Logger
	Client kube.IClient
}

func (r *Router) Mount(g *gin.RouterGroup) {
	g.POST("", r.auth(), func(c *gin.Context) {
		var webhook HarborWebhook
		err := c.BindJSON(&webhook)
		if err != nil {
			r.Logger.Info("JSON unmarshal failure", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"ok": false})
			return
		}
		if webhook.EventType == "PUSH_ARTIFACT" {
			err := r.handlePushArtifact(&webhook.EventData)
			if err != nil {
				r.Logger.Info("Error while updating deployment", zap.Error(err))
				c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"ok": false})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}

func (r *Router) handlePushArtifact(w *HarborWebhookEvent) error {
	r.Logger.Info("Received Harbor webhook", zap.String("image_name", w.Repository.FullName))
	for _, m := range r.Config.Mappings {
		if w.Repository.FullName == m.ImageName {
			err := r.Client.UpdateDeploymentImage(m.Namespace, m.DeploymentName, w.Resources[0].ResourceURL)
			r.Logger.Info("Updated deployment",
				zap.String("image_name", m.ImageName),
				zap.String("deployment", m.DeploymentName))
			return err
		}
	}
	return nil
}

func (r *Router) auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		headerValue := c.Request.Header.Get("Authorization")
		if headerValue == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		token := headerValue[len("Bearer "):]
		if token != r.Config.AuthToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
