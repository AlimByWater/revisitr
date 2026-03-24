package files

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"revisitr/internal/controller/http/middleware"
	"revisitr/internal/usecase/storage"
)

const maxFileSize = 50 << 20 // 50MB

type Group struct {
	storage   storage.Storage
	bucket    string
	jwtSecret string
}

func New(s storage.Storage, bucket, jwtSecret string) *Group {
	return &Group{storage: s, bucket: bucket, jwtSecret: jwtSecret}
}

func (g *Group) Path() string {
	return "/api/v1/files"
}

func (g *Group) Auth() gin.HandlerFunc {
	return middleware.Auth(g.jwtSecret)
}

func (g *Group) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleUpload,
	}
}

func (g *Group) handleUpload() (string, string, gin.HandlerFunc) {
	return http.MethodPost, "/upload", func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
			return
		}
		defer file.Close()

		if header.Size > maxFileSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("file too large, max %d MB", maxFileSize>>20),
			})
			return
		}

		ext := filepath.Ext(header.Filename)
		key := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

		info, err := g.storage.Upload(
			c.Request.Context(),
			g.bucket,
			key,
			file,
			header.Size,
			header.Header.Get("Content-Type"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
			return
		}

		c.JSON(http.StatusOK, info)
	}
}
