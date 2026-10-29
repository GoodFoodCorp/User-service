// Package avatarstore implements domain.AvatarStore on S3-compatible object
// storage (AWS S3, MinIO, Scaleway...) and, for local development only, on
// the local disk.
package avatarstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ── S3-compatible object storage ────────────────────────────

// S3Config locates the bucket. PublicURL is the base URL images are served
// from (bucket website, CDN in front of it...).
type S3Config struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
	PublicURL string
}

type S3Store struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewS3Store(cfg S3Config) (*S3Store, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &S3Store{client: client, bucket: cfg.Bucket, publicURL: strings.TrimRight(cfg.PublicURL, "/")}, nil
}

func (s *S3Store) Put(ctx context.Context, key, contentType string, body io.Reader, size int64) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
		// Keys are unique per upload: the image never changes behind its URL.
		CacheControl: "public, max-age=31536000, immutable",
	})
	if err != nil {
		return "", fmt.Errorf("store avatar: %w", err)
	}
	return s.publicURL + "/" + key, nil
}

func (s *S3Store) Delete(ctx context.Context, url string) error {
	key, ok := strings.CutPrefix(url, s.publicURL+"/")
	if !ok {
		return nil // not one of ours (legacy local URL)
	}
	return s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
}

// ── Local disk (development, single replica) ────────────────

// DiskStore writes under dir and serves the files at URLPrefix (see the
// router). Files live on one pod's disk: not for production.
type DiskStore struct {
	Dir       string
	URLPrefix string
}

func (d DiskStore) Put(_ context.Context, key, _ string, body io.Reader, _ int64) (string, error) {
	name := filepath.Base(key)
	if err := os.MkdirAll(d.Dir, 0o755); err != nil {
		return "", err
	}
	dst, err := os.Create(filepath.Join(d.Dir, name))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, body); err != nil {
		return "", err
	}
	return d.URLPrefix + name, nil
}

func (d DiskStore) Delete(_ context.Context, url string) error {
	name, ok := strings.CutPrefix(url, d.URLPrefix)
	if !ok || name == "" || strings.ContainsAny(name, `/\`) {
		return nil
	}
	if err := os.Remove(filepath.Join(d.Dir, name)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
