package avatarstore

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/minio/minio-go/v7"
)

func TestS3StorePutServeDelete(t *testing.T) {
	backend := s3mem.New()
	srv := httptest.NewServer(gofakes3.New(backend).Server())
	defer srv.Close()
	endpoint := strings.TrimPrefix(srv.URL, "http://")
	cfg := S3Config{Endpoint: endpoint, Bucket: "avatars", AccessKey: "k", SecretKey: "s", PublicURL: srv.URL + "/avatars"}
	store, err := NewS3Store(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := store.client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
		t.Fatal(err)
	}

	body := []byte("\x89PNG\r\n\x1a\nimage")
	url, err := store.Put(ctx, "avatars/u1-abc.png", "image/png", bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}
	if url != srv.URL+"/avatars/avatars/u1-abc.png" {
		t.Fatalf("url %q", url)
	}
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	served, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !bytes.Equal(served, body) {
		t.Fatalf("served %d %q", resp.StatusCode, served)
	}

	if err := store.Delete(ctx, url); err != nil {
		t.Fatal(err)
	}
	if _, err := store.client.StatObject(ctx, cfg.Bucket, "avatars/u1-abc.png", minio.StatObjectOptions{}); err == nil {
		t.Fatal("deleted avatar still stored")
	}
	if err := store.Delete(ctx, "/api/users/avatars/legacy.png"); err != nil {
		t.Fatal("foreign URLs are ignored:", err)
	}
}

func TestDiskStoreStaysInItsDirectory(t *testing.T) {
	dir := t.TempDir()
	store := DiskStore{Dir: dir, URLPrefix: "/api/users/avatars/"}

	url, err := store.Put(context.Background(), "avatars/../../etc/u1.png", "image/png", strings.NewReader("x"), 1)
	if err != nil {
		t.Fatal(err)
	}
	if url != "/api/users/avatars/u1.png" {
		t.Fatalf("url %q", url)
	}
	if _, err := os.Stat(filepath.Join(dir, "u1.png")); err != nil {
		t.Fatal("file not written in its directory")
	}
	if err := store.Delete(context.Background(), "/api/users/avatars/../secret"); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(context.Background(), url); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "u1.png")); !os.IsNotExist(err) {
		t.Fatal("file not deleted")
	}
}
