package tenablesc

import (
	"fmt"
)

const scanResultImportEndpoint = "/scanResult/import"

// ScanResultImport represents the request/response structure from the import endpoint of https://docs.tenable.com/tenablesc/api/Scan-Result.htm
type ScanResultImport struct {
	Filename             string   `json:"filename,omitempty"`
	Repository           BaseInfo `json:"repository,omitempty"`
	ClassifyMitigatedAge string   `json:"classifyMitigatedAge,omitempty"`
	DHCPTracking         FakeBool `json:"dhcpTracking,omitempty"`
	ScanningVirtualHosts FakeBool `json:"scanningVirtualHosts,omitempty"`
	RolloverType         string   `json:"rolloverType,omitempty"`
	TimeoutAction        string   `json:"timeoutAction,omitempty"`
}

func (c *Client) ImportScanResult(resultImport *ScanResultImport) error {

	var resp interface{} // Guess what! the response object is _empty_!

	_, err := c.postResource(scanResultImportEndpoint, resultImport, resp)

	return err
}

// ImportScanResultFile composes the UploadFile and ImportScanResult calls necessary to import a scan result to a target repository.
//
//	Path is expected to be a path on the local filesystem.
func (c *Client) ImportScanResultFile(path string, repositoryID string, classifyMitigatedAge string, dhcpTracking, scanningVirtualHosts bool) error {

	file, err := c.UploadFile(path, "")

	if err != nil {
		return fmt.Errorf("failed to upload file for scan result: %w", err)
	}

	sri := &ScanResultImport{
		Filename:             file.Filename,
		Repository:           BaseInfo{ID: ProbablyString(repositoryID)},
		ClassifyMitigatedAge: classifyMitigatedAge,
		DHCPTracking:         ToFakeBool(dhcpTracking),
		ScanningVirtualHosts: ToFakeBool(scanningVirtualHosts),
	}

	return c.ImportScanResult(sri)
}
