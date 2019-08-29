/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package legacyregistry

import (
	"bufio"
	"compress/gzip"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

/*
 * This file is mostly a copy of prometheus/http.go and has been backported here for
 * the purposes of retaining backwards compatibility. This file and the legacyregistry
 * are both deprecated, in favor of using explicitly defined local registries.
 */

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

var instLabels = []string{"method", "code"}

type nower interface {
	Now() time.Time
}

type nowFunc func() time.Time

func (n nowFunc) Now() time.Time {
	return n()
}

var now nower = nowFunc(func() time.Time {
	return time.Now()
})

// InstrumentedHandler works like InstrumentHandlerFunc (and shares
// the same issues) but provides more flexibility (at the cost of a more complex
// call syntax). See InstrumentHandlerWithOpts for details how the provided
// SummaryOpts are used.
//
// Deprecated: InstrumentHandlerFuncWithOpts is deprecated for the same reasons
// as InstrumentHandler is. Use the tooling provided in package promhttp instead.
func InstrumentedHandler(handlerFunc func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	opts := prometheus.SummaryOpts{
		Subsystem:   "http",
		ConstLabels: prometheus.Labels{"handler": "prometheus"},
		Objectives:  map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}
	reqCnt := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   opts.Namespace,
			Subsystem:   "http",
			Name:        "requests_total",
			Help:        "Total number of HTTP requests made.",
			ConstLabels: prometheus.Labels{"handler": "prometheus"},
		},
		instLabels,
	)
	if err := RawRegister(reqCnt); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			reqCnt = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	opts.Name = "request_duration_microseconds"
	opts.Help = "The HTTP request latencies in microseconds."
	reqDur := prometheus.NewSummary(opts)
	if err := RawRegister(reqDur); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			reqDur = are.ExistingCollector.(prometheus.Summary)
		} else {
			panic(err)
		}
	}

	opts.Name = "request_size_bytes"
	opts.Help = "The HTTP request sizes in bytes."
	reqSz := prometheus.NewSummary(opts)
	if err := RawRegister(reqSz); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			reqSz = are.ExistingCollector.(prometheus.Summary)
		} else {
			panic(err)
		}
	}

	opts.Name = "response_size_bytes"
	opts.Help = "The HTTP response sizes in bytes."
	resSz := prometheus.NewSummary(opts)
	if err := RawRegister(resSz); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			resSz = are.ExistingCollector.(prometheus.Summary)
		} else {
			panic(err)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		delegate := &responseWriterDelegator{ResponseWriter: w}
		out := computeApproximateRequestSize(r)

		_, cn := w.(http.CloseNotifier)
		_, fl := w.(http.Flusher)
		_, hj := w.(http.Hijacker)
		_, rf := w.(io.ReaderFrom)
		var rw http.ResponseWriter
		if cn && fl && hj && rf {
			rw = &fancyResponseWriterDelegator{delegate}
		} else {
			rw = delegate
		}
		handlerFunc(rw, r)

		elapsed := float64(time.Since(now)) / float64(time.Microsecond)

		method := sanitizeMethod(r.Method)
		code := sanitizeCode(delegate.status)
		reqCnt.WithLabelValues(method, code).Inc()
		reqDur.Observe(elapsed)
		resSz.Observe(float64(delegate.written))
		reqSz.Observe(float64(<-out))
	}
}

func computeApproximateRequestSize(r *http.Request) <-chan int {
	// Get URL length in current goroutine for avoiding a race condition.
	// HandlerFunc that runs in parallel may modify the URL.
	s := 0
	if r.URL != nil {
		s += len(r.URL.String())
	}

	out := make(chan int, 1)

	go func() {
		s += len(r.Method)
		s += len(r.Proto)
		for name, values := range r.Header {
			s += len(name)
			for _, value := range values {
				s += len(value)
			}
		}
		s += len(r.Host)

		// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

		if r.ContentLength != -1 {
			s += int(r.ContentLength)
		}
		out <- s
		close(out)
	}()

	return out
}

type responseWriterDelegator struct {
	http.ResponseWriter

	status      int
	written     int64
	wroteHeader bool
}

func (r *responseWriterDelegator) WriteHeader(code int) {
	r.status = code
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseWriterDelegator) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.written += int64(n)
	return n, err
}

type fancyResponseWriterDelegator struct {
	*responseWriterDelegator
}

func (f *fancyResponseWriterDelegator) CloseNotify() <-chan bool {
	return f.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (f *fancyResponseWriterDelegator) Flush() {
	f.ResponseWriter.(http.Flusher).Flush()
}

func (f *fancyResponseWriterDelegator) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return f.ResponseWriter.(http.Hijacker).Hijack()
}

func (f *fancyResponseWriterDelegator) ReadFrom(r io.Reader) (int64, error) {
	if !f.wroteHeader {
		f.WriteHeader(http.StatusOK)
	}
	n, err := f.ResponseWriter.(io.ReaderFrom).ReadFrom(r)
	f.written += n
	return n, err
}

func sanitizeMethod(m string) string {
	switch m {
	case "GET", "get":
		return "get"
	case "PUT", "put":
		return "put"
	case "HEAD", "head":
		return "head"
	case "POST", "post":
		return "post"
	case "DELETE", "delete":
		return "delete"
	case "CONNECT", "connect":
		return "connect"
	case "OPTIONS", "options":
		return "options"
	case "NOTIFY", "notify":
		return "notify"
	default:
		return strings.ToLower(m)
	}
}

func sanitizeCode(s int) string {
	switch s {
	case 100:
		return "100"
	case 101:
		return "101"

	case 200:
		return "200"
	case 201:
		return "201"
	case 202:
		return "202"
	case 203:
		return "203"
	case 204:
		return "204"
	case 205:
		return "205"
	case 206:
		return "206"

	case 300:
		return "300"
	case 301:
		return "301"
	case 302:
		return "302"
	case 304:
		return "304"
	case 305:
		return "305"
	case 307:
		return "307"

	case 400:
		return "400"
	case 401:
		return "401"
	case 402:
		return "402"
	case 403:
		return "403"
	case 404:
		return "404"
	case 405:
		return "405"
	case 406:
		return "406"
	case 407:
		return "407"
	case 408:
		return "408"
	case 409:
		return "409"
	case 410:
		return "410"
	case 411:
		return "411"
	case 412:
		return "412"
	case 413:
		return "413"
	case 414:
		return "414"
	case 415:
		return "415"
	case 416:
		return "416"
	case 417:
		return "417"
	case 418:
		return "418"

	case 500:
		return "500"
	case 501:
		return "501"
	case 502:
		return "502"
	case 503:
		return "503"
	case 504:
		return "504"
	case 505:
		return "505"

	case 428:
		return "428"
	case 429:
		return "429"
	case 431:
		return "431"
	case 511:
		return "511"

	default:
		return strconv.Itoa(s)
	}
}
