package tenablesc

import (
	"fmt"
)

type baseError struct {
	message string
	parent  error
}

func (b baseError) Error() string {
	return b.message
}

func (b baseError) Unwrap() error {
	return b.parent
}

type HTTPError struct {
	baseError
	ResponseCode int
	Body         string
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("%s, response code '%d' body:'%s'", h.message, h.ResponseCode, h.Body)
}

type NotFoundError HTTPError

type SCError struct {
	baseError
	SCErrorCode int
}

func (s SCError) Error() string {
	return fmt.Sprintf("%s, error code %d", s.message, s.SCErrorCode)
}
