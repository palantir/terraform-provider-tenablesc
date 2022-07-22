package tenablesc

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
)

const filesEndpoint = "/file"

// File represents the response structure for https://docs.tenable.com/tenablesc/api/File.htm
type File struct {
	Filename         string `json:"filename,omitempty"`
	OriginalFilename string `json:"originalFilename,omitempty"`
	Content          string `json:"content,omitempty"`
	BenchmarkName    string `json:"benchmarkName,omitempty"`
	ProfileName      string `json:"profileName,omitempty"`
	Version          string `json:"version,omitempty"`
	Type             string `json:"type,omitempty"`
}

func (c *Client) UploadFile(path, context string) (*File, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		//to make err check happy
		_ = file.Close()
	}()

	return c.UploadFileFromReader(file, path, context)
}

func (c *Client) UploadFileFromString(content, path, context string) (*File, error) {
	return c.UploadFileFromReader(strings.NewReader(content), path, context)
}

func (c *Client) UploadFileFromReader(reader io.Reader, path, context string) (*File, error) {

	bodyBuffer := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuffer)
	part, err := writer.CreateFormFile("Filedata", filepath.Base(path))
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart form: %w", err)
	}

	_, err = io.Copy(part, reader)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	f := &File{}

	req := c.client.NewRequest().
		SetBody(bodyBuffer).
		SetHeader("Content-Type", writer.FormDataContentType()).
		SetQueryParam("context", context)

	_, err = c.handleRequest(resty.MethodPost, fmt.Sprintf("%s/%s", filesEndpoint, "upload"), req, f)

	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return f, nil
}

func (c *Client) DeleteFile(filename string) error {
	f := &File{
		Filename: filename,
	}
	resp := &SCResponse{}
	req := c.client.NewRequest().SetBody(f).SetResult(resp).SetError(resp)
	r, err := req.Execute(resty.MethodPost, fmt.Sprintf("%s/%s", filesEndpoint, "clear"))

	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", filename, err)
	}

	if r.StatusCode() < 200 || r.StatusCode() > 299 {
		return fmt.Errorf("got response code %d when deleting file %s", r.StatusCode(), filename)
	}

	if resp.ErrorCode != 0 {
		return fmt.Errorf("got error code %d (%s) when deleting file %s", resp.ErrorCode, resp.ErrorMsg, filename)
	}

	return nil
}
