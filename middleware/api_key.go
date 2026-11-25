package Middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)


func ApiKey(key string) gin.HandlerFunc {
	expected := sha256.Sum256([]byte(key))

	return func(c *gin.Context) {
		provided := sha256.Sum256([]byte(c.GetHeader("Menu-API-Key")))

		if key == "" || subtle.ConstantTimeCompare(
			provided[:], expected[:],
		) != 1 {
			log.Printf("Tentative d'accès non autorisé : %s (%s)", c.ClientIP(), time.Now())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error" : "Clé API absente ou invalide",
			})
			return
		}
		c.Next()
	}
}
