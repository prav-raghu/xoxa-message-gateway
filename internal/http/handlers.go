// Package http provides the REST API for the message gateway.
package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"

	"xoxa-message-gateway/internal/config"
	"xoxa-message-gateway/internal/service"
	"xoxa-message-gateway/pkg/dto"
)

const swaggerIndexHTML = `<!DOCTYPE html>
<html>
  <head><title>xoxa-gateway API docs</title></head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = () => {
        window.ui = SwaggerUIBundle({
          url: '/docs/openapi.yaml',
          dom_id: '#swagger-ui',
        });
      };
    </script>
  </body>
</html>`

// RegisterHandlers wires every HTTP route onto r.
func RegisterHandlers(r *gin.Engine, svc *service.Service, cfg *config.Config) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")
	r.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerIndexHTML))
	})

	v1 := r.Group("/api/v1", AuthMiddleware(cfg))
	v1.POST("/messages", sendHandler(svc))
	v1.GET("/messages/:id", getMessageHandler(svc))
}

func sendHandler(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.SendRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resp, err := svc.SendMessage(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, resp)
	}
}

func getMessageHandler(svc *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		resp, err := svc.GetMessage(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
