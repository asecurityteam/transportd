package components

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type httpError struct {
	// Code is the HTTP status code.
	Code int `json:"code"`
	// Status is the HTTP status string.
	Status string `json:"status"`
	// Reason is the debug data.
	Reason string `json:"reason"`
}

func newError(code int, reason string) *http.Response {
	b, _ := json.Marshal(httpError{
		Code:   code,
		Status: http.StatusText(code),
		Reason: reason,
	})
	return &http.Response{
		Status:     http.StatusText(code),
		StatusCode: code,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       ioutil.NopCloser(bytes.NewReader(b)),
	}
}
