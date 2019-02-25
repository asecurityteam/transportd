package components

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/asecurityteam/runhttp"
)

type accessLog struct {
	Time                   string   `logevent:"@timestamp"`
	UGCDirty               []string `logevent:"ugc_dirty"`
	Schema                 string   `logevent:"schema,default=access"`
	SourceIP               string   `logevent:"src_ip"`
	ForwardedFor           string   `logevent:"forwarded_for"`
	DestinationIP          string   `logevent:"dest_ip"`
	Site                   string   `logevent:"site"`
	HTTPRequestContentType string   `logevent:"http_request_content_type"`
	HTTPMethod             string   `logevent:"http_method"`
	HTTPReferrer           string   `logevent:"http_referrer"`
	HTTPUserAgent          string   `logevent:"http_user_agent"`
	URIPath                string   `logevent:"uri_path"`
	URIQuery               string   `logevent:"uri_query"`
	Scheme                 string   `logevent:"scheme"`
	Port                   int      `logevent:"port"`
	Bytes                  int      `logevent:"bytes"`
	BytesOut               int      `logevent:"bytes_out"`
	BytesIn                int      `logevent:"bytes_in"`
	Duration               int      `logevent:"duration"`
	HTTPContentType        string   `logevent:"http_content_type"`
	Status                 int      `logevent:"status"`
	Message                string   `logevent:"message,default=access"`
}

type loggingTransport struct {
	Wrapped http.RoundTripper
}

// RoundTrip writes structured access logs for the request.
func (c *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var srcIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	var dstIP, dstPortStr, _ = net.SplitHostPort(r.Context().Value(http.LocalAddrContextKey).(net.Addr).String())
	var dstPort, _ = strconv.Atoi(dstPortStr)
	var a = accessLog{
		Time:                   time.Now().UTC().Format(time.RFC3339Nano),
		DestinationIP:          dstIP,
		Port:                   dstPort,
		SourceIP:               srcIP,
		Site:                   r.Host,
		HTTPRequestContentType: r.Header.Get("Content-Type"),
		HTTPMethod:             r.Method,
		HTTPReferrer:           r.Referer(),
		HTTPUserAgent:          r.UserAgent(),
		URIPath:                r.URL.Path,
		URIQuery:               r.URL.Query().Encode(),
		Scheme:                 r.URL.Scheme,
	}
	var start = time.Now()
	var resp, e = c.Wrapped.RoundTrip(r)
	a.Duration = int(time.Since(start).Nanoseconds() / 1e6)
	a.Status = -1
	if e == nil {
		a.Status = resp.StatusCode
		a.HTTPContentType = resp.Header.Get("Content-Type")
	}
	runhttp.LoggerFromContext(r.Context()).Info(a)
	return resp, e
}

// AccessLogConfig modifies the behavior of the access logs.
type AccessLogConfig struct{}

// Name of the config root.
func (*AccessLogConfig) Name() string {
	return "accesslog"
}

// AccessLogComponent is a logging plugin.
type AccessLogComponent struct{}

// AccessLog satisfies the NewComponent signature.
func AccessLog(_ context.Context, _ string, _ string, _ string) (interface{}, error) {
	return &AccessLogComponent{}, nil
}

// Settings generates a config populated with defaults.
func (m *AccessLogComponent) Settings() *AccessLogConfig {
	return &AccessLogConfig{}
}

// New generates the middleware.
func (*AccessLogComponent) New(ctx context.Context, conf *AccessLogConfig) (func(http.RoundTripper) http.RoundTripper, error) {
	return func(next http.RoundTripper) http.RoundTripper {
		return &loggingTransport{Wrapped: next}
	}, nil
}
