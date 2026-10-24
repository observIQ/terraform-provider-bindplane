// Copyright  observIQ, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// defaultTimeout is the request timeout, including reading the
	// response body. It matches the Bindplane client.
	defaultTimeout = 20 * time.Second

	headerAPIKey    = "X-Bindplane-Api-Key"    // #nosec G101 this is a header name, not a credential
	headerAccountID = "X-Bindplane-Account-ID" // #nosec G101 this is a header name, not a credential

	userAgent = "terraform-provider-bindplane"

	// statusPoliciesNotAccepted is returned by Bindplane when the user
	// has not accepted the required policies.
	statusPoliciesNotAccepted = 453

	// maxErrorBodySize limits how much of an error response is read.
	maxErrorBodySize = 1 << 20
)

// do sends a request to the Bindplane API. body, if not nil, is sent as JSON.
// A successful JSON response is decoded into result, if result is not nil.
// message describes the request in returned errors.
func (i *BindPlane) do(ctx context.Context, method, path string, query url.Values, body, result any, message string) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("%s: marshal request: %w", message, err)
		}
		reqBody = bytes.NewReader(data)
	}

	u := i.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return fmt.Errorf("%s: %w", message, err)
	}

	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	switch {
	case i.config.APIKey != "":
		req.Header.Set(headerAPIKey, i.config.APIKey)
	case i.config.Username != "" && i.config.Password != "":
		req.SetBasicAuth(i.config.Username, i.config.Password)
	}

	if i.config.AccountID != "" {
		req.Header.Set(headerAccountID, i.config.AccountID)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return i.statusError(resp, message)
	}

	// Bindplane serves its UI for unknown paths, which happens
	// when the server is older than the API used by the client.
	if strings.HasPrefix(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		return fmt.Errorf("%s: endpoint not found", message)
	}

	if result == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("%s: decode response: %w", message, err)
	}
	return nil
}

// statusError returns an error for an unsuccessful response, including
// the errors reported by Bindplane in the response body.
func (i *BindPlane) statusError(resp *http.Response, message string) error {
	if resp.StatusCode == statusPoliciesNotAccepted {
		return fmt.Errorf("policies not accepted. Log in to Bindplane at %s to accept the required policies before using the API", i.config.RemoteURL)
	}

	var errResp errorResponse
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
	if err != nil || json.Unmarshal(data, &errResp) != nil || len(errResp.Errors) == 0 {
		return fmt.Errorf("%s, got %s", message, resp.Status)
	}

	return fmt.Errorf("%s, got %s: %s", message, resp.Status, strings.Join(errResp.Errors, "\n"))
}

// get decodes the resource at path into result. It does not return an
// error if the resource does not exist, and result is left unchanged.
func (i *BindPlane) get(path string, result any) error {
	err := i.do(context.Background(), http.MethodGet, path, nil, nil, result, "unable to get "+path)
	if err != nil && isNotFoundError(err) {
		return nil
	}
	return err
}

// delete deletes the resource at path.
func (i *BindPlane) delete(path string) error {
	query := url.Values{"force": {"false"}}
	return i.do(context.Background(), http.MethodDelete, path, query, nil, nil, "unable to delete "+path)
}

// resourcePath returns the path of the named resource in collection.
func resourcePath(collection, name string) string {
	return "/" + collection + "/" + url.PathEscape(name)
}
