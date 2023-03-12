package direct

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.b8s.dev/rollingpin/config"
	"go.b8s.dev/rollingpin/kube"
	"go.uber.org/zap"
)

type DirectWebhook struct {
	ImageURL       string `json:"image_url"`
	RepositoryName string `json:"repository_name"`
}

type Router struct {
	Config *config.Config
	Logger *zap.Logger
	Client kube.IClient
}

func (r *Router) Mount(g *gin.RouterGroup) {
	g.POST("", r.auth(), func(c *gin.Context) {
		var webhook DirectWebhook
		err := c.BindJSON(&webhook)
		if err != nil {
			r.Logger.Info("JSON unmarshal failure", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"ok": false})
			return
		}
		err = r.handleWebhook(&webhook)
		if err != nil {
			r.Logger.Info("Error while updating deployment", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"ok": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
}

func (r *Router) handleWebhook(w *DirectWebhook) error {
	r.Logger.Info("Received direct webhook", zap.String("repository_name", w.RepositoryName))
	for _, m := range r.Config.Mappings {
		if w.RepositoryName == m.ImageName {
			err := r.Client.UpdateDeploymentImage(m.Namespace, m.DeploymentName, w.ImageURL)
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
