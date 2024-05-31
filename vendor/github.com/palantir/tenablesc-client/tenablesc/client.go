package tenablesc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	DefaultUserAgent = "tenable.sc go client"
)

type Client struct {
	client resty.Client
}

type response struct {
	response *resty.Response
}

// NewClient creates a Tenable.SC client object with defaults applied.
//
//	Don't forget to SetAPIKey or SetBasicAuth to ensure you make credentialed queries.
func NewClient(baseURL string) *Client {
	client := resty.New().
		SetBaseURL(baseURL).
		SetHeader(http.CanonicalHeaderKey("User-Agent"), DefaultUserAgent).
		AddRetryCondition(defaultTenableRetryConditions)

	return &Client{*client}
}

// SetAPIKey adds the API Key header to all queries with the client.
func (c *Client) SetAPIKey(access, secret string) *Client {
	c.client.SetHeader("x-apikey",
		fmt.Sprintf("accesskey=%s; secretkey=%s;",
			access,
			secret))
	return c
}

// SetBasicAuth adds the deprecated username/password auth to queries with the client.
func (c *Client) SetBasicAuth(username, password string) *Client {
	c.client.SetBasicAuth(username, password)
	return c
}

// SetUserAgent applies a UserAgent header; if this is not supplied DefaultUserAgent is used.
func (c *Client) SetUserAgent(agent string) *Client {
	c.client.SetHeader(http.CanonicalHeaderKey("User-Agent"), agent)
	return c
}

// RestyClient returns a pointer to the underlying resty.Client instance.
// This enables access to all the features and options provided by the resty library.
func (c *Client) RestyClient() *resty.Client {
	return &c.client
}

func defaultTenableRetryConditions(resp *resty.Response, err error) bool {

	// Assume internal server errors, gateway errors, and such are probably transient.
	if resp.StatusCode() >= 500 {
		return true
	}

	if resp.IsError() {
		// Some errors SC emits are transient issues like database locks.
		// for those things where waiting and trying again will suffice, return true.
		// At the moment, we don't have a clear 'retryable' flag to work with from the
		// vendor's opaque bitfield of an error code.
		if scr, ok := resp.Error().(SCResponse); ok {
			if strings.Contains(scr.ErrorMsg, "database is locked") {
				return true
			}
		}
	}

	// parse the response, if error and isRetryableSCError, yes.

	return false
}

// Tenable.SC's server expects all but the default set of fields to be specified as part of queries.
// This function inspects the provided interface for which fields should be requested.
// All `json` field names are included in the list;
// If the field includes `tenable:recurse` tag, then the child structure is also interrogated for
//
//	additional fields to extract.
func getFieldsForStruct(d interface{}) []string {
	t := reflect.TypeOf(d)

	//if a reflect.Type is passed in directly
	if typ, ok := d.(reflect.Type); ok {
		t = typ
	}
	for t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	fMap := map[string]interface{}{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if f := field.Tag.Get("tenable"); strings.Contains(f, "recurse") {
			for _, k := range getFieldsForStruct(field.Type) {
				fMap[k] = nil
			}
		} else if f := field.Tag.Get("json"); f != "" {
			fMap[strings.Split(f, ",")[0]] = nil
		} else if field.Anonymous {
			for _, k := range getFieldsForStruct(field.Type) {
				fMap[k] = nil
			}
		}
	}

	fields := make([]string, 0, len(fMap))

	for k := range fMap {
		fields = append(fields, k)
	}

	sort.Stable(sort.StringSlice(fields))

	return fields
}

// Generalized handlers for all endpoint queries.

func (c *Client) getResource(endpoint string, dest interface{}) (*response, error) {
	if !isPTR(dest) {
		return nil, errors.New("provide a pointer to the data source")
	}

	req := c.client.NewRequest()

	f := getFieldsForStruct(dest)
	if len(f) > 0 {
		req.SetQueryParam("fields",
			strings.Join(f, ","))
	}

	return c.handleRequest(resty.MethodGet, endpoint, req, dest)
}

func (c *Client) postResource(endpoint string, input interface{}, dest interface{}) (*response, error) {
	if !isPTR(dest) {
		return nil, errors.New("provide a pointer to the data source")
	}

	req := c.client.NewRequest().SetBody(input)

	return c.handleRequest(resty.MethodPost, endpoint, req, dest)
}

func (c *Client) patchResource(endpoint string, input interface{}, dest interface{}) (*response, error) {
	if !isPTR(dest) {
		return nil, errors.New("provide a pointer to the data source")
	}

	req := c.client.NewRequest().SetBody(input)

	return c.handleRequest(resty.MethodPatch, endpoint, req, dest)
}

func (c *Client) patchResourceWithID(endpoint string, input interface{}, dest interface{}) (*response, error) {

	id, err := idFromStruct(input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ID from struct: %w", err)
	}
	endpoint = fmt.Sprintf("%s/%s", endpoint, id)

	return c.patchResource(endpoint, input, dest)
}

func (c *Client) deleteResource(endpoint string, input interface{}, dest interface{}) (*response, error) {
	if !isPTR(dest) {
		return nil, errors.New("provide a pointer to the data source")
	}

	req := c.client.NewRequest().SetBody(input)

	return c.handleRequest(resty.MethodDelete, endpoint, req, dest)
}

// handleRequest implements the application-side retry and backoff logic for all queries, retrying in case of
//
//	application-side errors that are clearly transient.
func (c *Client) handleRequest(method, endpoint string, request *resty.Request, dest interface{}) (*response, error) {
	var err error

	if request == nil {
		request = c.client.NewRequest()
	}

	req := request.
		// Whether a good response or an error, give us the SCResponse parse.
		SetResult(SCResponse{}).
		SetError(SCResponse{})

	resp, err := req.Execute(method, endpoint)

	if err != nil {
		return &response{resp}, fmt.Errorf("failed to make request: %w", err)
	}

	return &response{resp}, handleResponse(resp, dest)

}

type SCResponse struct {
	Response  json.RawMessage `json:"response"`
	ErrorCode int             `json:"error_code"`
	ErrorMsg  string          `json:"error_msg"`
	Timestamp int             `json:"timestamp"`
	Warnings  []string        `json:"warning"`
}

func handleHTTPError(resp *resty.Response) error {
	var respErr error
	if resp.StatusCode() < 200 || resp.StatusCode() > 299 {

		httpErr := HTTPError{
			baseError: baseError{
				message: "unexpected response from server",
			},
			ResponseCode: resp.StatusCode(),
			Body:         string(resp.Body()),
		}

		//SC's version of not found for some reason.
		if resp.StatusCode() == 403 {
			e := NotFoundError(httpErr)
			e.baseError.parent = httpErr
			respErr = e
		} else {
			respErr = httpErr
		}
		return respErr
	}

	return nil
}

// handleResponse's job is to handle finishing the unmarshal, as well as
//
//	wrapping the error if there's an error here.
func handleResponse(resp *resty.Response, dest interface{}) error {
	respErr := handleHTTPError(resp)

	//try to unmarshal the response anyways incase there's something interesting
	scr := &SCResponse{}
	if err := json.Unmarshal(resp.Body(), scr); err != nil {
		if respErr != nil {
			return respErr
		}
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if scr.ErrorCode != 0 {
		return SCError{
			baseError: baseError{
				message: scr.ErrorMsg,
				parent:  respErr,
			},
			SCErrorCode: scr.ErrorCode,
		}
	}
	if dest != nil {
		if err := json.Unmarshal(scr.Response, dest); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func isPTR(d interface{}) bool {
	t := reflect.TypeOf(d)

	return d == nil || t.Kind() == reflect.Ptr
}

func idFromStruct(d interface{}) (string, error) {
	v := reflect.ValueOf(d)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", errors.New("method only functions on structs or struct pointers")
	}

	id := v.FieldByName("ID")

	if !id.IsValid() {
		return "", errors.New("no ID field was found")
	} else if id.Kind() != reflect.String {
		return "", errors.New("id field is not a string")
	} else if id.IsZero() {
		return "", errors.New("id field is empty")
	}

	return id.String(), nil
}
