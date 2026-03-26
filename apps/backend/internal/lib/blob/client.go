package blob

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type Client struct {
	container *azblob.Client
	baseURL   string
}

func New(connectionString, containerName, accountName string) (*Client, error) {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("create blob client: %w", err)
	}

	baseURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName)
	return &Client{container: client, baseURL: baseURL}, nil
}

func (c *Client) UploadPoster(ctx context.Context, containerName, movieID string, data []byte, contentType string) (string, error) {
	ext := extensionFromContentType(contentType)
	name := fmt.Sprintf("%s%s", movieID, ext)
	_, err := c.container.UploadStream(ctx, containerName, name, bytes.NewReader(data), nil)
	if err != nil {
		return "", fmt.Errorf("upload poster: %w", err)
	}
	return c.baseURL + "/" + path.Base(name), nil
}

func extensionFromContentType(contentType string) string {
	switch strings.ToLower(contentType) {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
