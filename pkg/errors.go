package transportd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func newError(code int, reason string) *http.Response {
	b, _ := json.Marshal(HTTPError{
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
