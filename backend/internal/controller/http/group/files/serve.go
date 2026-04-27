package files

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type fileServer interface {
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, string, error)
}

// StorageServe serves files from MinIO storage. Registered as a separate group
// at /revisitr/storage to match the URL pattern used in the database.
type StorageServe struct {
	fs     fileServer
	bucket string
}

func NewStorageServe(fs fileServer, bucket string) *StorageServe {
	return &StorageServe{fs: fs, bucket: bucket}
}

func (g *StorageServe) Path() string {
	return "/revisitr/storage"
}

func (g *StorageServe) Auth() gin.HandlerFunc {
	return nil // public access
}

func (g *StorageServe) Handlers() []func() (string, string, gin.HandlerFunc) {
	return []func() (string, string, gin.HandlerFunc){
		g.handleServe,
	}
}

func (g *StorageServe) handleServe() (string, string, gin.HandlerFunc) {
	return http.MethodGet, "/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.Status(http.StatusBadRequest)
			return
		}

		reader, contentType, err := g.fs.GetObject(c.Request.Context(), g.bucket, key)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer reader.Close()

		if contentType == "" {
			contentType = "application/octet-stream"
		}
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Header("Content-Type", contentType)
		c.Status(http.StatusOK)
		io.Copy(c.Writer, reader)
	}
}
