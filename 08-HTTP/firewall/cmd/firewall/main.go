//go:build !solution

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
)

func main() {
	conf := flag.String("conf", "", "Path to config with rules (required)")
	serviceAddr := flag.String("service-addr", "", "Address of the main server (required)")
	addr := flag.String("addr", "", "Address of the firewall server (required)")

	flag.Parse()

	rules, err := parseYaml(*conf)
	if err != nil {
		log.Fatal(err)
	}

	err = rules.check()
	if err != nil {
		log.Fatal(err)
	}

	target, err := url.Parse(*serviceAddr)
	if err != nil {
		log.Fatalf("invalid upstream address: %v", err)
	}

	if target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") {
		log.Fatal("service-addr must be an absolute HTTP or HTTPS URL")
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		server(w, r, target, rules)
	}

	if err = http.ListenAndServe(*addr, http.HandlerFunc(handler)); err != nil {
		log.Fatal(err)
	}
}

func server(w http.ResponseWriter, r *http.Request, target *url.URL, rules *rules) {
	var firewallRule *rule

	endpoint := r.URL.Path
	for _, rule := range rules.Rules {
		if endpoint == rule.Endpoint {
			firewallRule = &rule
			break
		}
	}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(target)
			pr.SetXForwarded()
		},
		ModifyResponse: func(resp *http.Response) error {
			return responseCheck(resp, firewallRule)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Println(err)

			if errors.Is(err, errForbidden) {
				writeForbidden(w)
				return
			}

			http.Error(w, "Bad Gateway", http.StatusBadGateway)
		},
	}

	err := requestChecks(r, firewallRule)
	if err != nil {
		log.Println(err)
		writeForbidden(w)

		return
	}

	proxy.ServeHTTP(w, r)
}

var errForbidden = errors.New("forbidden by firewall rule")

func writeForbidden(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	_, _ = io.WriteString(w, "Forbidden")
}

func matchAny(patterns []*regexp.Regexp, value []byte) bool {
	for _, pattern := range patterns {
		if pattern.Match(value) {
			return true
		}
	}

	return false
}

func requestChecks(r *http.Request, rule *rule) error {
	if rule == nil {
		return nil
	}

	if matchAny(rule.userAgentPatterns, []byte(r.UserAgent())) {
		return fmt.Errorf("ForbiddenUserAgents: %w", errForbidden)
	}

	for name, values := range r.Header {
		for _, value := range values {
			if matchAny(rule.headerPatterns, []byte(name+": "+value)) {
				return fmt.Errorf("ForbiddenHeaders: %w", errForbidden)
			}
		}
	}

	for _, header := range rule.RequiredHeaders {
		if _, ok := r.Header[http.CanonicalHeaderKey(header)]; !ok {
			return fmt.Errorf("missing required header %q: %w", header, errForbidden)
		}
	}

	body, err := inspectBody(r.Body, rule.MaxRequestLengthBytes, rule.requestPatterns)
	if err == nil {
		r.Body = body
	}

	return err
}

func responseCheck(resp *http.Response, rule *rule) error {
	if rule == nil {
		return nil
	}

	for _, code := range rule.ForbiddenResponseCodes {
		if code == resp.StatusCode {
			return fmt.Errorf("ForbiddenResponseCodes: %w", errForbidden)
		}
	}

	body, err := inspectBody(resp.Body, rule.MaxResponseLengthBytes, rule.responsePatterns)
	if err == nil {
		resp.Body = body
	}

	return err
}

func inspectBody(body io.ReadCloser, limit *int, patterns []*regexp.Regexp) (io.ReadCloser, error) {
	if limit == nil && len(patterns) == 0 {
		return body, nil
	}

	if body == nil {
		body = http.NoBody
	}

	const maxInspectBytes = 1024 * 1024

	maxBytes := maxInspectBytes
	if limit != nil {
		maxBytes = *limit
		if len(patterns) > 0 {
			maxBytes = min(maxBytes, maxInspectBytes)
		}
	}

	reader := io.Reader(body)
	if int64(maxBytes) < math.MaxInt64 {
		reader = io.LimitReader(body, int64(maxBytes)+1)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if len(data) > maxBytes {
		return nil, fmt.Errorf("body exceeds size limit: %w", errForbidden)
	}

	if matchAny(patterns, data) {
		return nil, fmt.Errorf("body matches forbidden pattern: %w", errForbidden)
	}

	_ = body.Close()

	return io.NopCloser(bytes.NewReader(data)), nil
}
