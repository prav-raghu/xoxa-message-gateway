package http

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"xoxa-message-gateway/internal/config"
)

// AuthMiddleware validates a Bearer JWT on every request, using HS256 if
// cfg.JWTSecret is configured, otherwise an RSA public key from disk.
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	var rsaKey *rsa.PublicKey
	if cfg.JWTSecret == "" && cfg.JWTPublicKey != "" {
		if keyBytes, err := os.ReadFile(cfg.JWTPublicKey); err == nil {
			if key, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes); err == nil {
				rsaKey = key
			}
		}
	}

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			switch t.Method.(type) {
			case *jwt.SigningMethodHMAC:
				if cfg.JWTSecret == "" {
					return nil, errors.New("HS256 not configured")
				}
				return []byte(cfg.JWTSecret), nil
			case *jwt.SigningMethodRSA:
				if rsaKey == nil {
					return nil, errors.New("RS256 not configured")
				}
				return rsaKey, nil
			default:
				return nil, errors.New("unsupported signing method")
			}
		},
			jwt.WithIssuer(cfg.JWTIssuer),
			jwt.WithAudience(cfg.JWTAudience),
		)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
