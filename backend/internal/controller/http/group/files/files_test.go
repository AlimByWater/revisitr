package files

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"revisitr/internal/usecase/storage"
)

type mockStorage struct {
	uploadFn func(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error)
}

func (m *mockStorage) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error) {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, bucket, key, reader, size, contentType)
	}
	return &storage.FileInfo{Key: key, URL: "/files/" + key, ContentType: contentType, Size: size}, nil
}

func (m *mockStorage) Delete(context.Context, string, string) error           { return nil }
func (m *mockStorage) GetURL(context.Context, string, string) (string, error) { return "", nil }

func newFilesTestEngine(g *Group) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	rg := engine.Group(g.Path())
	for _, fn := range g.Handlers() {
		method, path, handler := fn()
		rg.Handle(method, path, handler)
	}
	return engine
}

func newMultipartRequest(t *testing.T, filename string, content []byte) (*http.Request, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if filename != "" {
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("write form file: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, writer.FormDataContentType()
}

func TestHandleUpload_RequiresFile(t *testing.T) {
	g := New(&mockStorage{}, "bucket", "secret")
	engine := newFilesTestEngine(g)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/files/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "file required") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleUpload_RejectsTooLargeFile(t *testing.T) {
	g := New(&mockStorage{}, "bucket", "secret")
	engine := newFilesTestEngine(g)

	payload := bytes.Repeat([]byte("a"), maxFileSize+1)
	req, _ := newMultipartRequest(t, "huge.bin", payload)
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "file too large") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleUpload_ReturnsInternalErrorWhenStorageFails(t *testing.T) {
	g := New(&mockStorage{
		uploadFn: func(_ context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error) {
			return nil, errors.New("boom")
		},
	}, "bucket", "secret")
	engine := newFilesTestEngine(g)

	req, _ := newMultipartRequest(t, "avatar.png", []byte("png-data"))
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "upload failed") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleUpload_ReturnsUploadedFileInfo(t *testing.T) {
	var seenBucket, seenContentType string
	var seenSize int64
	g := New(&mockStorage{
		uploadFn: func(_ context.Context, bucket, key string, reader io.Reader, size int64, contentType string) (*storage.FileInfo, error) {
			seenBucket = bucket
			seenContentType = contentType
			seenSize = size
			if !strings.HasSuffix(key, ".png") {
				t.Fatalf("expected key with .png suffix, got %q", key)
			}
			return &storage.FileInfo{Key: key, URL: "/uploads/" + key, ContentType: contentType, Size: size}, nil
		},
	}, "uploads", "secret")
	engine := newFilesTestEngine(g)

	req, _ := newMultipartRequest(t, "avatar.png", []byte("png-data"))
	req.Header.Set("Content-Type", req.Header.Get("Content-Type"))
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if seenBucket != "uploads" {
		t.Fatalf("expected bucket uploads, got %q", seenBucket)
	}
	if seenContentType == "" {
		t.Fatal("expected content type to be passed through")
	}
	if seenSize != int64(len("png-data")) {
		t.Fatalf("expected size %d, got %d", len("png-data"), seenSize)
	}
	if !strings.Contains(w.Body.String(), "\"url\":\"/uploads/") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
