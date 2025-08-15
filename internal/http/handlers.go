// Package http provides HTTP handlers for the API
package http

import (
	"github.com/gin-gonic/gin"
	"xoxa-message-gateway/internal/service"
	"xoxa-message-gateway/pkg/dto"
)

func RegisterHandlers(r *gin.Engine) {
	r.POST("/send", sendHandler)
}

func sendHandler(c *gin.Context) {
	var req dto.SendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	resp := service.SendMessage(req)
	c.JSON(200, resp)
}
