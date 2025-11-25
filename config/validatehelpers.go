/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
)

// validateURL parses a URL and returns either the parsed URL or an error.
func validateURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, errors.New("URL is empty")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid URL %q: %w", raw, err)
	}
	return u, nil
}

// validateHostPort ensures the URL contains a valid host:port pair.
func validateHostPort(u *url.URL) error {
	if u == nil {
		return errors.New("URL is nil")
	}

	// Some schemes (rist/udp/rtp/srt) expect host:port
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return fmt.Errorf("invalid host:port in URL %q: %w", u.String(), err)
	}
	if host == "" || port == "" {
		return fmt.Errorf("missing host or port in URL %q", u.String())
	}
	return nil
}

// validateUnique checks that a list of inputs/outputs contains no duplicates
// based on Identifier OR URL OR URL.Host.
func validateUnique(list interface{}, name string) error {
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("validateUnique expects a slice, got %T", list)
	}

	check := make(map[string]bool)

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()

		key := ""
		var u *url.URL
		var err error

		switch obj := item.(type) {

		case Input:
			key = obj.Identifier
			u, err = validateURL(obj.URL)
			if err != nil {
				return fmt.Errorf("invalid input URL in %s: %w", name, err)
			}

		case Output:
			key = obj.Identifier
			u, err = validateURL(obj.URL)
			if err != nil {
				return fmt.Errorf("invalid output URL in %s: %w", name, err)
			}

		default:
			return fmt.Errorf("validateUnique: unsupported type %T", item)
		}

		// Check duplicates by identifier
		if check[key] {
			return fmt.Errorf("duplicate identifier %q in %s", key, name)
		}
		check[key] = true

		// Check duplicates by host (127.0.0.1:5000)
		if u != nil {
			if check[u.Host] {
				return fmt.Errorf("duplicate host %q in %s", u.Host, name)
			}
			check[u.Host] = true
		}
	}

	return nil
}
