<a id="markdown-transportd---http-middleware-as-a-service" name="transportd---http-middleware-as-a-service"></a>
# transportd - HTTP Middleware As A Service
[![GoDoc](https://godoc.org/github.com/asecurityteam/transportd?status.svg)](https://godoc.org/github.com/asecurityteam/transportd)
[![Build Status](https://travis-ci.com/asecurityteam/transportd.png?branch=master)](https://travis-ci.com/asecurityteam/transportd)
[![codecov.io](https://codecov.io/github/asecurityteam/transportd/coverage.svg?branch=master)](https://codecov.io/github/asecurityteam/transportd?branch=master)

*Status: Incubation*

<!-- TOC -->

- [transportd - HTTP Middleware As A Service](#transportd---http-middleware-as-a-service)
    - [Overview](#overview)
    - [Configuration](#configuration)
        - [Runtime Settings](#runtime-settings)
        - [Backend Settings](#backend-settings)
        - [Route Settings](#route-settings)
        - [Environment Variables](#environment-variables)
    - [Custom Plugins And Builds](#custom-plugins-and-builds)
        - [Custom Components](#custom-components)
        - [Writing A Component](#writing-a-component)
        - [Generating A Build](#generating-a-build)
    - [Using As A Library](#using-as-a-library)
    - [Contributing](#contributing)
        - [License](#license)
        - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-overview" name="overview"></a>
## Overview

This project aggregates all of our most effective tooling into an HTTP reverse proxy
that removes all direct references to our tools from application code and extends
the possible languages/technical stacks that make use of our tools. The list of
features we offer includes:

-   Support for timeout, retries, and tail-end latency correction.
-   Highly detailed performance metrics output and structured access logging.
-   Request and response validation based on the OpenAPI specification.
-   A simple JWT based authentication layer built on our open
    [ASAP](https://github.com/asecurityteam/asap) standard.
-   (In Development) Cascading failure protection through our
    curve based [load shedding](https://github.com/asecurityteam/loadshed) algorithm.

This is a companion piece to our HTTP client middleware projects at
[transport](https://github.com/asecurityteam/transport),
[httpstats](https://github.com/asecurityteam/httpstats), and
[loadshed](https://github.com/asecurityteam/loadshed) that wraps those libraries in
a runtime of their own that can be maintained outside of the main application code.

We continue to maintain those core tools as libraries for higher performing, more
resilient, and better operated HTTP client calls in go. Over the years, however,
we've started to push the boundaries of what a common set of libraries can do
and have started targeting a common runtime layer to help ease of use. Most notably,
we've run into two primary problems: 1) our go systems end up with large sections of
copy/paste to import, configure, and install the tools which leads to subtle skew
between projects and 2) we are unable to share the advancements we've made with our
libraries outside of systems written in go.

Note that this project is meant to fulfill a relatively simple need of adapting our
existing libraries to more services. We don't intend for this project to become a
fully featured "smart proxy" or "service mesh" appliance. While there is a decent
overlap in features, we won't focus on service discovery or multi-protocol support.
For those kinds of features we recommend [Netflix's Zuul](https://github.com/Netflix/zuul)
with [Netflix's Eureka](https://github.com/Netflix/eureka) or
[Lyft's Envoy](https://www.envoyproxy.io/).

<a id="markdown-configuration" name="configuration"></a>
## Configuration

This proxy is built around OpenAPI consumes an OpenAPI specification as configuration.
It looks for a set of extensions at certain key points in the document. Below are
each of the locations and extensions the system expects in order to run.

<a id="markdown-runtime-settings" name="runtime-settings"></a>
### Runtime Settings

The HTTP server is provided by another one of our libraries,
[runhttp](https://github.com/asecurityteam/runhttp), and comes with its own extension
section in the OpenAPI specification. The system will look for a top-level extension
called `x-runtime` and use the contents to configure the server:

```yaml
x-runtime:
  signals:
    # ([]string) Which signal handlers are installed. Choices are OS.
    installed:
      - "OS"
    os:
      # ([]int) Which signals to listen for.
      signals:
        - 15
        - 2
  stats:
    # (string) Destination stream of the stats. One of NULLSTAT, DATADOG.
    output: "DATADOG"
    datadog:
      # (int) Max packet size to send.
      packetsize: 32768
      # ([]string) Any static tags for all metrics.
      tags:
      # (time.Duration) Frequencing of sending metrics to listener.
      flushinterval: "10s"
      # (string) Listener address to use when sending metrics.
      address: "localhost:8125"
  logger:
    # (string) Destination stream of the logs. One of STDOUT, NULL.
    output: "STDOUT"
    # (string) The minimum level of logs to emit. One of DEBUG, INFO, WARN, ERROR.
    level: "INFO"
  httpserver:
    # (string) The listening address of the server.
    address: ":8080"
```

<a id="markdown-backend-settings" name="backend-settings"></a>
### Backend Settings

In addition to the runtime, the system also expects that all backend configurations
will be present in a top-level extensions called `x-transportd`. Backend configuration
is what links the reverse proxy to one or more other systems. Each backend is
configured with a host name for the destination and a connection pool setting:

```yaml
x-transportd:
  # ([]string) Available backends. Names are symbolic and referenced later.
  backends:
    - "backendName"
  backendName:
    # (string) Backend host URL.
    host: "https://localhost"
    pool:
      # (int) Number of connections pools. Only use >1 if HTTP/2
      count: 1
      # (time.Duration) Lifetime of a pool before refreshing.
      ttl: "1h0m0s"
```

<a id="markdown-route-settings" name="route-settings"></a>
### Route Settings

Each route in the OpenAPI specification will need its own `x-transportd`
extension block in order to configure the specific client behaviors that
you want. Each block must contain the relevant backend reference for routing,
an ordered list of installed middleware, and the middleware specific settings
for any installed:

```yaml
x-transportd:
  # ([]string) Ordered list of components enabled for this route.
  enabled:
    - "metrics"
    - "accesslog"
    - "asapvalidate"
    - "validateheaders"
    - "timeout"
    - "hedging"
    - "retry"
    - "asaptoken"
    - "requestvalidation"
    - "responsevalidation"
    - "strip"
    - "requestheaderinject"
    - "responseheaderinject"
    - "basicauth"
  # (string) Backend target for this route.
  backend: "backendName"
  metrics:
    # (string) Name of the tag containing the path referecne.
    pathtag: "client_path"
    # (string) Name of the tag containing the backend reference.
    backendtag: "client_dependency"
    # (string) Name of idle connection return count metric.
    putidle: "http.client.put_idle"
    # (string) Name of bytes sent and received metric.
    bytestotal: "http.client.bytes_total"
    # (string) Name of bytes sent metric.
    bytessent: "http.client.bytes_sent"
    # (string) Name of bytes received metric.
    bytesreceived: "http.client.bytes_received"
    # (string) Name of time to first resposne byte metrics.
    firstresponsebyte: "http.client.first_response_byte.timing"
    # (string) Name of time to write headers metric.
    wroteheaders: "http.client.wrote_headers.timing"
    # (string) Name of TLS timing metric.
    tls: "http.client.tls.timing"
    # (string) Name of idle timing metric.
    connectionidle: "http.client.connection_idle.timing"
    # (string) Name of TCP timing metric.
    tcp: "http.client.tcp.timing"
    # (string) Name of DNS timing metric.
    dns: "http.client.dns.timing"
    # (string) Name of overall timing metric.
    timing: "http.client.timing"
  asapvalidate:
    # ([]string) Public key download URLs.
    keyurls:
    # (string) Acceptable audience string.
    allowedaudience: ""
    # ([]string) Acceptable issuer strings.
    allowedissuers:
  validateheaders:
    # (map[string][]string) allowed list of headers whose values to check
    allowed:
      accept:
        - "text/json"
        - "text/html"
    # (map[string] string) the delimiters to use for splitting header-specific values when they come in single line
    split:
      accept: ","
  timeout:
    # (time.Duration) Duration after which the request is canceled.
    after: "175ms"
  hedging:
    # (time.Duration) Duration after which to open a new request.
    interval: "50ms"
  retry:
    # (time.Duration) Time to wait between requests.
    backoff: "50ms"
    # (int) Maximum retry attempts.
    limit: 3
    # ([]int) HTTP status codes that trigger a retry.
    codes:
      - 500
      - 501
      - 502
      - 503
      - 504
      - 505
      - 506
      - 507
      - 508
      - 509
      - 510
      - 511
  asaptoken:
    # ([]string) JWT audience values to include in tokens.
    audiences:
    # (string) JWT issuer value to include in tokens.
    issuer: ""
    # (time.Duration) Lifetime of a token.
    ttl: "0s"
    # (string) JWT kid value to include in tokens.
    kid: ""
    # (string) RSA private key to use when signing tokens.
    privatekey: ""
  strip:
    # (int) Number of URL segments to remove from the beginning of the path before redirect.
    count: 0
  requestheaderinject:
    # (map[string][]string) Map values of header names:values to inject.
    headers:
      x-header-1:
        - "value1"
      x-header-2:
        - "value2"
  responseheaderinject:
    # (map[string][]string) Map values of header names:values to inject.
    headers:
      x-header-1:
        - "value1"
      x-header-2:
        - "value2"
  basicauth:
    # (string) Username to use in HTTP basic authentication.
    username: ""
    # (string) Password to use in HTTP basic authentication.
    password: ""
```

<a id="markdown-environment-variables" name="environment-variables"></a>
### Environment Variables

For cases where a static YAML file is insufficient, such as deploying to multiple
environments or regions that each require slightly different configurations,
we also offer support for using environment variable references in the OpenAPI
specification. Before validating the document and loading configuration, the
system will look for the pattern `${}` and replace all instances with the
environment variable value that is identified by the name inside the pattern.
For example, `${FOO}` will result in the `FOO` environment variable being fetched
and the value inserted.

<a id="markdown-custom-plugins-and-builds" name="custom-plugins-and-builds"></a>
## Custom Plugins And Builds

Unfortunately, go does not have good support for dynamic loading of plugins and the
alternatives to the standard library `plugin` package all require some form of
multi-processing and RPC. Because of this, adding features beyond the core set
requires creating a custom build of this project.

<a id="custom-components" name="custom-components"></a>
### Custom Components

Documentation for custom components that handle things such as header validation can be found at [docs/components.md](docs/components.md)


<a id="markdown-writing-a-component" name="writing-a-component"></a>
### Writing A Component

This project uses another one of our libraries, [settings](https://github.com/asecurityteam/settings),
to manage configuration and plugins. The full suite of what can be done with a
component is available in the `settings` documentation. For convenience, here's
a summary:

Plugins do not need to directly reference this project nor do they need to directly
reference the `settings` project. All new components must have three basic features:
1) a configuration struct defined, 2) an implementation of the `settings.Component`
interface, and 3) a route-aware constructor for the component. The best way to
demonstrate this is with an existing plugin. Here is the annotated source for the
`timeout` plugin:

```golang
package myplugin

import (
	"context"
	"net/http"
	"time"
)

const (
	defaultTimeoutSettingAfter = 175 * time.Millisecond
)

// All core functionality should be built as a decorator for the
// http.RoundTripper interface. The goal is to layer on functionality
// without the need for tight-coupling between components.
//
// The general pattern for a decorator is a struct that takes in the
// wrapped http.RoundTripper and exposes its own RoundTrip method.
type timeoutRoundTripper struct {
	Wrapped http.RoundTripper
	after time.Duration
}

// RoundTrip is the method required to satisfy the http.RoundTripper interface
// and allows your decorator to appear as though it is an underlying HTTP client.
// Your decorator can do just about anything it needs to so long as there is at
// least one path that results in the wrapped http.RoundTripper being called.
func (m *timeoutRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	var ctx, cancel = context.WithTimeout(r.Context(), m.after)
	defer cancel()
	return m.Wrapped.RoundTrip(r.WithContext(ctx))
}

// All configurations are defined as structs. Each field of the struct will
// equate to one setting in the configuration. The names of the individual
// settings are bound to the struct field names.
//
// You can add an optional description annotation to struct fields which will
// display when printing help text.
type TimeoutConfig struct {
	After time.Duration `description:"Duration after which the request is canceled."`
}

// You can also define an optional Name() method for the struct that will change
// the key used to identify the root of the configuration section. By default,
// the key will match the struct name. For example, if we remove this method
// then the system will expect configuration in the form of:
//
//    timeoutconfig:
//      after: "200ms"
//
// Versus with the method:
//
//    timeout:
//      after: "200ms"
func (*TimeoutConfig) Name() string {
	return "timeout"
}

// Each component also needs a bit of scaffolding to work with our plugin
// system. The scaffolding must be a struct but the struct is not required
// to actually maintain any kind of state in the form of struct fields.
// If desired, though, you can use the struct to store details such as
// the active HTTP path or method which will be provided in the constructor.
type TimeoutComponent struct{}

// Each component struct needs a constructor function in order to be installed
// in the list of plugins. The constructor function signature is exported from
// the project as transportd.NewComponent and matches the method below. If your
// component needs to be aware of any of the values provided then you may store
// those values on your component struct.
func Timeout(ctx context.Context, backend string, path string, method string) (interface{}, error) {
	return &TimeoutComponent{}, nil
}

// As part of the plugin system, your component struct must define a method
// called 'Settings()' that returns an pointer to and instance of your configuration
// struct. The instance should come populated with any default values which will be
// used both as the defaults and as the displayed value in help text.
func (*TimeoutComponent) Settings() *TimeoutConfig {
	return &TimeoutConfig{After: defaultTimeoutSettingAfter}
}

// The final piece of your component is a constructor function for the middleware you
// want installed. The constructor function must be called 'New', must take a context.Context
// as the first argument, and must take the same type in the second argument as returned
// by the 'Settings()' method. The return value must match this example.
func (*TimeoutComponent) New(_ context.Context, conf *TimeoutConfig) (func(http.RoundTripper) http.RoundTripper, error) { // nolint
  // the context object is guaranteed to have a pointer to the raw openapi3.Swagger document.  If the component needs
  // this pointer, it should change the `_` to `ctx` in the named function parameters, and:
  // ctx.Value(transportd.ContextKeyOpenAPISpec.String("doc")).(*openapi3.Swagger)
	return func(next http.RoundTripper) http.RoundTripper {
		return &timeoutRoundTripper{RoundTripper: next, after: conf.After}
	}, nil
}
```

<a id="markdown-generating-a-build" name="generating-a-build"></a>
### Generating A Build

Creating a custom build is equivalent to copying the `main.go` from this project
into your own repository and adding any custom components you've built to the
list:

```golang
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	transportd "github.com/asecurityteam/transportd/pkg"
	"github.com/asecurityteam/transportd/pkg/components"
)

func main() {
	ctx := context.Background()
	plugins := []transportd.NewComponent{
		components.Metrics,
		components.AccessLog,
		components.ASAPValidate,
		components.ValidateHeaders,
		components.Timeout,
		components.Hedging,
		components.Retry,
		components.ASAPToken,
		components.RequestValidation,
		components.ResponseValidation,
		components.Strip,
		// Insert any custom components here.
		// The order doesn't matter because the installation order is
		// determined by each path configuration.
	}

	// Handle the -h flag and print settings.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {}
	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		help, errHelp := transportd.Help(ctx, plugins...)
		if errHelp != nil {
			panic(errHelp.Error())
		}
		fmt.Println(help)
		return
	}

	// The system will accept either a full OpenAPI specification through
	// the environment or the name of a file where the specification is
	// stored. Priority is given to the file if both are present.
	fileName := os.Getenv("TRANSPORTD_OPENAPI_SPECIFICATION_FILE")
	fileContent := []byte(os.Getenv("TRANPSPORTD_OPENAPI_SPECIFICATION_CONTENT"))
	var errRead error
	if fileName != "" {
		fileContent, errRead = ioutil.ReadFile(fileName)
		if errRead != nil {
			panic(errRead)
		}
	}

	// Create and run the system.
	rt, err := transportd.New(ctx, fileContent, plugins...)
	if err != nil {
		panic(err.Error())
	}
	if err := rt.Run(); err != nil {
		panic(err.Error())
	}
}
```

<a id="markdown-using-as-a-library" name="using-as-a-library"></a>
## Using As A Library

While the source projects like
[transport](https://github.com/asecurityteam/transport) and
[httpstats](https://github.com/asecurityteam/httpstats) can be used directly, there
is a great deal of convenience in being able to configure all of the combined tooling
through the OpenAPI extensions. This is particularly true if you are moving a system
from a container orchestration environment to one where running the proxy alongside
the application is not possible. To help in these cases we offer the following:

```golang
transport, err := transportd.NewTransport(ctx, fileContent, plugins...)
if err != nil {
  panic(err.Error())
}
client := &http.Client{
  Transport: transport,
}
```

The resulting `http.RoundTripper` implements all of the smart functionality of the
reverse proxy but is exposed as a component that can be embedded in code and used
anywhere an HTTP client would otherwise be used.

<a id="markdown-contributing" name="contributing"></a>
## Contributing

<a id="markdown-license" name="license"></a>
### License

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
