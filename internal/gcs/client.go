package gcs

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client represents a Google Cloud Storage client
type Client struct {
	client    *storage.Client
	ctx       context.Context
	projectID string
}

// Item represents a bucket, folder or object in GCS
type Item struct {
	Name      string
	Path      string
	FullPath  string
	Size      int64
	Updated   time.Time
	IsDir     bool
	IsBucket  bool
	ParentDir string
}

// NewClient creates a new GCS client
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %v", err)
	}

	return &Client{
		client:    client,
		ctx:       ctx,
		projectID: projectID,
	}, nil
}

// Close closes the GCS client
func (c *Client) Close() error {
	return c.client.Close()
}

// ListBuckets lists all buckets in the GCS project
func (c *Client) ListBuckets() ([]Item, error) {
	var items []Item

	// Use the project ID when listing buckets
	it := c.client.Buckets(c.ctx, c.projectID)

	for {
		bucketAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing buckets: %v", err)
		}

		items = append(items, Item{
			Name:     bucketAttrs.Name,
			Path:     bucketAttrs.Name,
			FullPath: bucketAttrs.Name,
			Updated:  bucketAttrs.Updated,
			IsDir:    true,
			IsBucket: true,
		})
	}

	return items, nil
}

// ListObjects lists objects in a bucket with the given prefix
func (c *Client) ListObjects(bucketName, prefix string) ([]Item, error) {
	var items []Item
	bucket := c.client.Bucket(bucketName)

	// If prefix doesn't end with a slash and is not empty, add a slash
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// Get parent directory
	parentDir := ""
	if prefix != "" {
		parts := strings.Split(strings.TrimSuffix(prefix, "/"), "/")
		if len(parts) > 0 {
			parentDir = strings.Join(parts[:len(parts)-1], "/")
		}
	}

	// Add a special item to navigate up a directory
	if prefix != "" {
		items = append(items, Item{
			Name:      "..",
			Path:      parentDir,
			FullPath:  path.Join(bucketName, parentDir),
			IsDir:     true,
			ParentDir: parentDir,
		})
	}

	// Create a map to track directories
	directories := make(map[string]bool)

	it := bucket.Objects(c.ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})

	// Process common prefixes (directories)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing objects: %v", err)
		}

		if attrs.Prefix != "" {
			// This is a directory
			dirName := path.Base(strings.TrimSuffix(attrs.Prefix, "/"))
			directories[dirName] = true
			items = append(items, Item{
				Name:      dirName,
				Path:      attrs.Prefix,
				FullPath:  path.Join(bucketName, attrs.Prefix),
				IsDir:     true,
				ParentDir: prefix,
			})
		} else {
			// This is a file
			fileName := path.Base(attrs.Name)
			dirPath := path.Dir(attrs.Name)
			if dirPath == "." {
				dirPath = ""
			}

			// Only include files in the current directory
			if dirPath == strings.TrimSuffix(prefix, "/") || (dirPath == "" && prefix == "") {
				items = append(items, Item{
					Name:      fileName,
					Path:      attrs.Name,
					FullPath:  path.Join(bucketName, attrs.Name),
					Size:      attrs.Size,
					Updated:   attrs.Updated,
					IsDir:     false,
					ParentDir: prefix,
				})
			}
		}
	}

	return items, nil
}

// GetObjectContent gets the content of an object as a string
func (c *Client) GetObjectContent(bucketName, objectName string) (string, error) {
	bucket := c.client.Bucket(bucketName)
	obj := bucket.Object(objectName)

	reader, err := obj.NewReader(c.ctx)
	if err != nil {
		return "", fmt.Errorf("error opening object: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("error reading object: %v", err)
	}

	return string(data), nil
}

// ParsePath parses a full path into bucket name and prefix
func ParsePath(fullPath string) (string, string) {
	parts := strings.SplitN(fullPath, "/", 2)

	bucketName := parts[0]
	prefix := ""

	if len(parts) > 1 {
		prefix = parts[1]
	}

	return bucketName, prefix
}
