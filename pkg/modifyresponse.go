package transportd

import (
	"net/http"
	"net/url"
)

// MultiResponseModifier enables composition of ModifyResponse functions.
type MultiResponseModifier []func(*http.Response) error

// ModifyResponse satisfies the signature of the same name in the ReverseProxy.
func (mrs MultiResponseModifier) ModifyResponse(resp *http.Response) error {
	for _, mr := range mrs {
		if err := mr(resp); err != nil {
			return err
		}
	}
	return nil
}

// EnforceRelativeLocation prevents redirection from the backend as the result
// of 3xx codes from improperly redirecting clients around the proxy. This is
// related to https://github.com/golang/go/issues/14237.
func EnforceRelativeLocation(resp *http.Response) error {
	loc := resp.Header.Get("Location")
	if loc == "" {
		return nil
	}
	u, err := url.Parse(loc)
	if err != nil {
		// If Location isn't a valid URL then we skip the logic as it isn't
		// needed.
		return nil
	}
	// Removing the host and scheme results in /path/to/api?queries=here
	// rather than the original https://hostname.com/path/to/api?queries=here.
	u.Host = ""
	u.Scheme = ""
	resp.Header.Set("Location", u.String())
	return nil
}
