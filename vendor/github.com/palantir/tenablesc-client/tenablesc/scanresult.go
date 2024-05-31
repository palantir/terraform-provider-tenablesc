package tenablesc

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	scanResultEndpoint = "/scanResult"
)

var DefaultTimeScope = time.Time{}

var pkzipFileSignature = []byte{'\x50', '\x4b', '\x03', '\x04'}

// ScanResult represents the request/response structure from https://docs.tenable.com/tenablesc/api/Scan-Result.htm
type ScanResult struct {
	BaseInfo
	Status                 string         `json:"status,omitempty"`
	Initiator              UserInfo       `json:"initiator,omitempty"`
	Owner                  UserInfo       `json:"owner,omitempty"`
	OwnerGroup             BaseInfo       `json:"ownerGroup,omitempty"`
	Repository             BaseInfo       `json:"repository,omitempty"`
	Scan                   BaseInfo       `json:"scan,omitempty"`
	ImportStatus           string         `json:"importStatus,omitempty"`
	ImportStart            ProbablyString `json:"importStart,omitempty"`
	ImportFinish           ProbablyString `json:"importFinish,omitempty"`
	ImportDuration         ProbablyString `json:"importDuration,omitempty"`
	DownloadFormat         string         `json:"downloadFormat,omitempty"`
	DataFormat             string         `json:"dataFormat,omitempty"`
	ResultType             string         `json:"resultType,omitempty"`
	ResultSource           string         `json:"resultSource,omitempty"`
	ErrorDetails           string         `json:"errorDetails,omitempty"`
	ImportErrorDetails     string         `json:"importErrorDetails,omitempty"`
	TotalIPs               ProbablyString `json:"totalIPs,omitempty"`
	ScannedIPs             ProbablyString `json:"scannedIPs,omitempty"`
	StartTime              ProbablyString `json:"startTime,omitempty"`
	FinishTime             ProbablyString `json:"finishTime,omitempty"`
	ScanDuration           ProbablyString `json:"scanDuration,omitempty"`
	CompletedIPs           ProbablyString `json:"completedIPs,omitempty"`
	CompletedChecks        ProbablyString `json:"completedChecks,omitempty"`
	TotalChecks            ProbablyString `json:"totalChecks,omitempty"`
	AgentScanUUID          string         `json:"agentScanUUID,omitempty"`
	AgentScanContainerUUID string         `json:"agentScanContainerUUID,omitempty"`
	Job                    BaseInfo       `json:"job,omitempty"`
	Details                string         `json:"details,omitempty"`
}

type scanResultInternal struct {
	Manageable []*ScanResult `json:"manageable" tenable:"recurse"`
	Usable     []*ScanResult `json:"usable" tenable:"recurse"`
}

// takes startTime + endTime parameters, but defaults to last 30d.

func (c *Client) GetAllScanResults() ([]*ScanResult, error) {
	return c.GetAllScanResultsByTime(DefaultTimeScope, DefaultTimeScope)
}

func (c *Client) GetAllScanResultsByTime(start, end time.Time) ([]*ScanResult, error) {

	params := url.Values{}
	if !start.IsZero() {
		params.Add("startTime", fmt.Sprintf("%d", start.Unix()))
	}
	if !end.IsZero() {
		params.Add("endTime", fmt.Sprintf("%d", end.Unix()))
	}

	resourceURL := url.URL{
		Path:     scanResultEndpoint,
		RawQuery: params.Encode(),
	}

	var resp scanResultInternal
	if _, err := c.getResource(resourceURL.String(), &resp); err != nil {
		return nil, fmt.Errorf("failed to get scan results: %w", err)
	}

	var spOut []*ScanResult
	spMap := make(map[ProbablyString]bool)

	for _, o := range resp.Usable {
		spOut = append(spOut, o)
		spMap[o.ID] = true
	}
	for _, o := range resp.Manageable {
		if _, exists := spMap[o.ID]; !exists {
			spOut = append(spOut, o)
			spMap[o.ID] = true
		}
	}

	return spOut, nil
}

func (c *Client) GetScanResult(id string) (*ScanResult, error) {
	var resp ScanResult
	if _, err := c.getResource(fmt.Sprintf("%s/%s", scanResultEndpoint, id), &resp); err != nil {
		return nil, fmt.Errorf("failed to get scan result %s: %w", id, err)
	}

	return &resp, nil
}

func (c *Client) DeleteScanResult(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", scanResultEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to delete scan result with id %s: %w", id, err)
	}

	return nil
}

// StopScanResult Stops the Scan Result associated with {id}.
// NOTE: This endpoint is not applicable for Agent Sync Results.
// ref: https://docs.tenable.com/tenablesc/api/Scan-Result.htm#ScanResultRESTReference-/scanResult/{id}/stop
func (c *Client) StopScanResult(id string) error {
	if _, err := c.postResource(fmt.Sprintf("%s/%s/stop", scanResultEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to stop scan result with id %s: %w", id, err)
	}

	return nil
}

func (c *Client) DownloadScanResult(id string) ([]byte, error) {

	possiblyZippedStream, err := c.internalDownloadScanResult(id)
	if err != nil {
		return nil, err
	}

	// so, it's _probably_ zipped, but the spec doesn't indicate when it is
	// and isn't, and I've gotten _both_! so.
	if !byteSliceIsPKZipped(possiblyZippedStream) {
		return possiblyZippedStream, nil
	}

	return firstFileFromPKZipSlice(possiblyZippedStream)

}

func (c *Client) internalDownloadScanResult(id string) ([]byte, error) {
	req := c.client.NewRequest()
	req.SetBody(struct {
		DownloadType string `json:"downloadType"`
	}{DownloadType: "v2"},
	)

	resp, err := req.Execute(resty.MethodPost,
		fmt.Sprintf("%s/%s/download", scanResultEndpoint, id),
	)
	if err != nil {
		return nil, err
	}

	if respErr := handleHTTPError(resp); respErr != nil {
		return nil, respErr
	}

	return resp.Body(), nil
}

func byteSliceIsPKZipped(slice []byte) bool {
	if len(slice) < len(pkzipFileSignature) {
		return false
	}
	for i, v := range pkzipFileSignature {
		if slice[i] != v {
			return false
		}
	}
	return true
}

func firstFileFromPKZipSlice(slice []byte) ([]byte, error) {

	var results []byte

	reader, err := zip.NewReader(bytes.NewReader(slice), int64(len(slice)))
	if err != nil {
		return nil, fmt.Errorf("failed to create zip reader: %w", err)
	}

	if len(reader.File) == 0 {
		return nil, errors.New("got empty zip for nessus scan result")
	}

	file, err := reader.Open(reader.File[0].Name)
	if err != nil {
		return nil, fmt.Errorf("could not open first file in zip: %w", err)
	}

	results, err = io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read zip file: %w", err)
	}

	return results, nil
}
