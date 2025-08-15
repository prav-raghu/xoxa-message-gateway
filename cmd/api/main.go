// Package main starts the HTTP server using Gin
package main

import (
	"github.com/gin-gonic/gin"
	"xoxa-message-gateway/internal/http"
)

func main() {
	r := gin.Default()
	http.RegisterHandlers(r)
	r.Run() // listen and serve on 0.0.0.0:8080
}
