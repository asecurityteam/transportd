package components

import (
	"context"
	"net/http"

	"github.com/asecurityteam/httpstats"
)

// MetricsConfig contains settings for request metrics emissions.
type MetricsConfig struct {
	Timing            string `description:"Name of overall timing metric."`
	DNS               string `description:"Name of DNS timing metric."`
	TCP               string `description:"Name of TCP timing metric."`
	ConnectionIdle    string `description:"Name of idle timing metric."`
	TLS               string `description:"Name of TLS timing metric."`
	WroteHeaders      string `description:"Name of time to write headers metric."`
	FirstResponseByte string `description:"Name of time to first resposne byte metrics."`
	BytesReceived     string `description:"Name of bytes received metric."`
	BytesSent         string `description:"Name of bytes sent metric."`
	BytesTotal        string `description:"Name of bytes sent and received metric."`
	PutIdle           string `description:"Name of idle connection return count metric."`
	BackendTag        string `description:"Name of the tag containing the backend reference."`
	PathTag           string `description:"Name of the tag containing the path referecne."`
}

// Name of the config root.
func (*MetricsConfig) Name() string {
	return "metrics"
}

// MetricsComponent implements the settings.Component interface.
type MetricsComponent struct {
	Backend string
	Path    string
}

// Metrics satisfies the NewComponent signature.
func Metrics(_ context.Context, backend string, path string, _ string) (interface{}, error) {
	return &MetricsComponent{Backend: backend, Path: path}, nil
}

// Settings generates a config populated with defaults.
func (*MetricsComponent) Settings() *MetricsConfig {
	return &MetricsConfig{
		Timing:            "http.client.timing",
		DNS:               "http.client.dns.timing",
		TCP:               "http.client.tcp.timing",
		ConnectionIdle:    "http.client.connection_idle.timing",
		TLS:               "http.client.tls.timing",
		WroteHeaders:      "http.client.wrote_headers.timing",
		FirstResponseByte: "http.client.first_response_byte.timing",
		BytesReceived:     "http.client.bytes_received",
		BytesSent:         "http.client.bytes_sent",
		BytesTotal:        "http.client.bytes_total",
		PutIdle:           "http.client.put_idle",
		BackendTag:        "client_dependency",
		PathTag:           "client_path",
	}
}

// New generates the middleware.
func (c *MetricsComponent) New(_ context.Context, conf *MetricsConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
	return httpstats.NewTransport(
		httpstats.TransportOptionBytesInName(conf.BytesReceived),
		httpstats.TransportOptionBytesOutName(conf.BytesSent),
		httpstats.TransportOptionBytesTotalName(conf.BytesTotal),
		httpstats.TransportOptionConnectionIdleName(conf.ConnectionIdle),
		httpstats.TransportOptionDNSName(conf.DNS),
		httpstats.TransportOptionFirstResponseByteName(conf.FirstResponseByte),
		httpstats.TransportOptionGotConnectionName(conf.TCP),
		httpstats.TransportOptionPutIdleName(conf.PutIdle),
		httpstats.TransportOptionRequestTimeName(conf.Timing),
		httpstats.TransportOptionTLSName(conf.TLS),
		httpstats.TransportOptionWroteHeadersName(conf.WroteHeaders),
		httpstats.TransportOptionTag(conf.BackendTag, c.Backend),
		httpstats.TransportOptionTag(conf.PathTag, c.Path),
	), nil
}
