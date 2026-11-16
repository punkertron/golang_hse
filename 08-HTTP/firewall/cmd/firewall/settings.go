//go:build !solution

package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type rules struct {
	Rules []rule `yaml:"rules"`
}

type rule struct {
	Endpoint               string   `yaml:"endpoint"`
	ForbiddenUserAgents    []string `yaml:"forbidden_user_agents"`
	ForbiddenHeaders       []string `yaml:"forbidden_headers"`
	RequiredHeaders        []string `yaml:"required_headers"`
	ForbiddenRequestRe     []string `yaml:"forbidden_request_re"`
	ForbiddenResponseRe    []string `yaml:"forbidden_response_re"`
	MaxRequestLengthBytes  *int     `yaml:"max_request_length_bytes"`
	MaxResponseLengthBytes *int     `yaml:"max_response_length_bytes"`
	ForbiddenResponseCodes []int    `yaml:"forbidden_response_codes,flow"`
	userAgentPatterns      []*regexp.Regexp
	headerPatterns         []*regexp.Regexp
	requestPatterns        []*regexp.Regexp
	responsePatterns       []*regexp.Regexp
}

func parseYaml(path string) (*rules, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := &rules{}

	err = yaml.NewDecoder(f).Decode(r)
	if errors.Is(err, io.EOF) {
		return r, nil
	}

	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *rules) check() error {
	for i := range r.Rules {
		rule := &r.Rules[i]
		if (rule.MaxRequestLengthBytes != nil && *rule.MaxRequestLengthBytes < 0) ||
			(rule.MaxResponseLengthBytes != nil && *rule.MaxResponseLengthBytes < 0) {
			return fmt.Errorf("endpoint %q: body size limits must be nonnegative", rule.Endpoint)
		}

		for _, field := range []struct {
			name     string
			patterns []string
			compiled *[]*regexp.Regexp
		}{
			{"forbidden_user_agents", rule.ForbiddenUserAgents, &rule.userAgentPatterns},
			{"forbidden_headers", rule.ForbiddenHeaders, &rule.headerPatterns},
			{"forbidden_request_re", rule.ForbiddenRequestRe, &rule.requestPatterns},
			{"forbidden_response_re", rule.ForbiddenResponseRe, &rule.responsePatterns},
		} {
			compiled := make([]*regexp.Regexp, 0, len(field.patterns))
			for _, pattern := range field.patterns {
				re, err := regexp.Compile(pattern)
				if err != nil {
					return fmt.Errorf("endpoint %q, %s: %w", rule.Endpoint, field.name, err)
				}

				compiled = append(compiled, re)
			}

			*field.compiled = compiled
		}
	}

	return nil
}
