package ig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	igcfg "github.com/felipeinf/instago/config"
	"github.com/felipeinf/instago/igerrors"
	"github.com/felipeinf/instago/password"
)

var publicLastMu sync.Mutex
var publicLastTS time.Time

func (c *Client) publicThrottle() {
	publicLastMu.Lock()
	defer publicLastMu.Unlock()
	if d := time.Since(publicLastTS); d < time.Second {
		time.Sleep(time.Second - d)
	}
	publicLastTS = time.Now()
}

// PublicRequest performs an HTTP call on httpPublic with light rate limiting; returnJSON unmarshals the body into any.
func (c *Client) PublicRequest(reqURL string, method string, body io.Reader, hdr http.Header, returnJSON bool) (any, error) {
	c.publicThrottle()
	if c.requestTimeout > 0 {
		time.Sleep(c.requestTimeout)
	}
	c.publicReqCount++
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}
	if hdr != nil {
		req.Header = hdr.Clone()
	}
	resp, err := c.httpPublic.Do(req)
	if err != nil {
		return nil, &igerrors.ClientConnection{ClientError: igerrors.ClientError{Message: err.Error()}}
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		if e := igerrors.MapPrivateHTTPError(reqURL, resp, raw); e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("public request %d", resp.StatusCode)
	}
	if returnJSON {
		var out any
		if err := json.Unmarshal(raw, &out); err != nil {
			if strings.Contains(reqURL, "/login/") {
				return nil, &igerrors.LoginRequired{ClientError: igerrors.ClientError{Message: err.Error()}}
			}
			return nil, err
		}
		return out, nil
	}
	return string(raw), nil
}

// PublicGraphqlRequest GETs www.instagram.com/graphql/query with variables JSON and query_hash, returning the data object.
func (c *Client) PublicGraphqlRequest(variables map[string]any, queryHash string) (map[string]any, error) {
	b, err := json.Marshal(variables)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("variables", buf.String())
	if queryHash != "" {
		params.Set("query_hash", queryHash)
	}
	u := igcfg.GraphQLPublicAPIURL + "?" + params.Encode()
	h := http.Header{}
	h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	out, err := c.PublicRequest(u, http.MethodGet, nil, h, true)
	if err != nil {
		return nil, err
	}
	root, ok := out.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("ig: graphql unexpected root")
	}
	if st, _ := root["status"].(string); st != "" && st != "ok" {
		return nil, fmt.Errorf("ig: graphql status %v msg %v", root["status"], root["message"])
	}
	data, _ := root["data"].(map[string]any)
	return data, nil
}

// WebProfileInfo calls the logged-out web API users/web_profile_info for username (lowercased).
func (c *Client) WebProfileInfo(username string) (map[string]any, error) {
	u := fmt.Sprintf("%sapi/v1/users/web_profile_info/?username=%s", igcfg.PublicWebURL, url.QueryEscape(strings.ToLower(username)))
	h := http.Header{}
	h.Set("x-ig-app-id", igcfg.DefaultIGAppID)
	h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	return c.publicJSONGet(u, h)
}

func (c *Client) publicJSONGet(u string, h http.Header) (map[string]any, error) {
	out, err := c.PublicRequest(u, http.MethodGet, nil, h, true)
	if err != nil {
		return nil, err
	}
	m, ok := out.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("ig: expected json object")
	}
	return m, nil
}

// FetchPasswordEncryptionKeys reads ig-set-password-encryption headers from the public client (same as password.FetchPublicKeys).
func (c *Client) FetchPasswordEncryptionKeys() (keyID int, pubKey string, err error) {
	return password.FetchPublicKeys(c.httpPublic)
}
