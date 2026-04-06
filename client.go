package ig

import (
	"bytes"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	igcfg "github.com/felipeinf/instago/config"
	igenc "github.com/felipeinf/instago/encoding"
	"github.com/felipeinf/instago/igerrors"
	"github.com/felipeinf/instago/internal/pickapp"
)

// Logger receives optional debug-style messages from the client; the default is a no-op.
type Logger interface {
	Infof(format string, args ...any)
	Debugf(format string, args ...any)
	Warnf(format string, args ...any)
}

type nopLogger struct{}

func (nopLogger) Infof(string, ...any)  {}
func (nopLogger) Debugf(string, ...any) {}
func (nopLogger) Warnf(string, ...any)  {}

// Client holds HTTP clients, session cookies, device emulation, and user caches for Instagram API calls.
type Client struct {
	httpPrivate *http.Client
	httpPublic  *http.Client
	logger      Logger

	username string
	password string

	uuids             UUIDs
	mid               string
	igURur            string
	igWWWClaim        string
	authData          map[string]string
	deviceSettings    DeviceSettings
	userAgent         string
	locale            string
	country           string
	countryCode       int
	timezoneOffset    int
	bloksVersioningID string
	appID             string

	// OverrideAppVersion, when true before LoadSettings, allows the client to replace the stored app version with a supported profile.
	OverrideAppVersion bool
	token              string

	requestTimeout time.Duration
	delayMin       time.Duration
	delayMax       time.Duration

	privateReqCount int
	publicReqCount  int
	lastLogin       *float64

	// LastHTTPResponse is the most recent private API HTTP response when available.
	LastHTTPResponse *http.Response
	// LastJSON is the decoded JSON body from the last private API response.
	LastJSON map[string]any

	rng *rand.Rand

	userMu      sync.Mutex
	usersByPK   map[int64]User
	usersByName map[string]int64
}

// NewClient returns a client with fresh device IDs, default locale, and empty session.
func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	jarPub, _ := cookiejar.New(nil)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	c := &Client{
		httpPrivate:    &http.Client{Jar: jar},
		httpPublic:     &http.Client{Jar: jarPub},
		logger:         nopLogger{},
		country:        "US",
		countryCode:    1,
		locale:         "en_US",
		timezoneOffset: -14400,
		appID:          igcfg.FBAnalyticsAppID,
		authData:       map[string]string{},
		rng:            rng,
		requestTimeout: time.Second,
		usersByPK:      map[int64]User{},
		usersByName:    map[string]int64{},
	}
	c.applyDefaultDevice()
	c.setUUIDs(UUIDs{})
	c.refreshUserAgent("")
	return c
}

// SetLogger sets structured log output; nil is ignored and the previous logger is kept.
func (c *Client) SetLogger(l Logger) {
	if l != nil {
		c.logger = l
	}
}

// SetProxy configures HTTP and HTTPS proxy for both private and public clients; empty clears the proxy.
func (c *Client) SetProxy(proxyURL string) error {
	if proxyURL == "" {
		c.httpPrivate.Transport = nil
		c.httpPublic.Transport = nil
		return nil
	}
	u, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}
	tr := &http.Transport{Proxy: http.ProxyURL(u)}
	c.httpPrivate.Transport = tr
	c.httpPublic.Transport = tr
	return nil
}

func (c *Client) applyDefaultDevice() {
	d := igcfg.DefaultDevice
	c.deviceSettings = DeviceSettings{
		AndroidVersion: d.AndroidVersion,
		AndroidRelease: d.AndroidRelease,
		DPI:            d.DPI,
		Resolution:     d.Resolution,
		Manufacturer:   d.Manufacturer,
		Device:         d.Device,
		Model:          d.Model,
		CPU:            d.CPU,
	}
	c.pickApp("")
}

func (c *Client) pickApp(seed string) {
	if c.OverrideAppVersion {
		p := pickapp.PickBySeed(seed, c.rng)
		c.deviceSettings.AppVersion = p.AppVersion
		c.deviceSettings.VersionCode = p.VersionCode
		c.deviceSettings.BloksVersioningID = p.BloksVersioningID
		c.bloksVersioningID = p.BloksVersioningID
		return
	}
	if p, ok := pickapp.MatchStored(c.deviceSettings.AppVersion); ok {
		c.deviceSettings.AppVersion = p.AppVersion
		c.deviceSettings.VersionCode = p.VersionCode
		c.deviceSettings.BloksVersioningID = p.BloksVersioningID
		c.bloksVersioningID = p.BloksVersioningID
		return
	}
	if c.deviceSettings.AppVersion != "" && !c.OverrideAppVersion {
		return
	}
	p := pickapp.PickBySeed(seed, c.rng)
	c.deviceSettings.AppVersion = p.AppVersion
	c.deviceSettings.VersionCode = p.VersionCode
	c.deviceSettings.BloksVersioningID = p.BloksVersioningID
	c.bloksVersioningID = p.BloksVersioningID
}

func (c *Client) refreshUserAgent(override string) {
	if override != "" {
		c.userAgent = override
		return
	}
	c.userAgent = strings.NewReplacer(
		"{app_version}", c.deviceSettings.AppVersion,
		"{android_version}", strconv.Itoa(c.deviceSettings.AndroidVersion),
		"{android_release}", c.deviceSettings.AndroidRelease,
		"{dpi}", c.deviceSettings.DPI,
		"{resolution}", c.deviceSettings.Resolution,
		"{manufacturer}", c.deviceSettings.Manufacturer,
		"{model}", c.deviceSettings.Model,
		"{device}", c.deviceSettings.Device,
		"{cpu}", c.deviceSettings.CPU,
		"{locale}", c.locale,
		"{version_code}", c.deviceSettings.VersionCode,
	).Replace(igcfg.UserAgentBase)
}

// SetLocale updates locale and derived country for headers and rebuilds the default User-Agent.
func (c *Client) SetLocale(locale string) {
	c.locale = locale
	if strings.Contains(locale, "_") {
		parts := strings.Split(locale, "_")
		if len(parts) > 1 {
			c.country = parts[len(parts)-1]
		}
	}
	c.refreshUserAgent("")
}

// SetTimezoneOffset sets the timezone offset in seconds sent with API payloads.
func (c *Client) SetTimezoneOffset(sec int) {
	c.timezoneOffset = sec
}

// SetDeviceSettings replaces emulated hardware fields and refreshes app version selection and User-Agent.
func (c *Client) SetDeviceSettings(d DeviceSettings) {
	c.deviceSettings = d
	c.pickApp(c.uuids.UUID)
	c.refreshUserAgent("")
}

// SetUserAgent sets a fixed User-Agent string for private requests (non-empty override).
func (c *Client) SetUserAgent(ua string) {
	c.refreshUserAgent(ua)
}

func (c *Client) setUUIDs(u UUIDs) {
	if u.PhoneID == "" {
		u.PhoneID = c.generateUUID("", "")
	}
	if u.UUID == "" {
		u.UUID = c.generateUUID("", "")
	}
	if u.ClientSessionID == "" {
		u.ClientSessionID = c.generateUUID("", "")
	}
	if u.AdvertisingID == "" {
		u.AdvertisingID = c.generateUUID("", "")
	}
	if u.AndroidDeviceID == "" {
		u.AndroidDeviceID = c.generateAndroidDeviceID()
	}
	if u.RequestID == "" {
		u.RequestID = c.generateUUID("", "")
	}
	if u.TraySessionID == "" {
		u.TraySessionID = c.generateUUID("", "")
	}
	c.uuids = u
}

// SetUUIDs replaces client identifiers and re-runs app profile selection for the new seed.
func (c *Client) SetUUIDs(u UUIDs) {
	c.setUUIDs(u)
	c.pickApp(u.UUID)
}

func (c *Client) generateUUID(prefix, suffix string) string {
	return prefix + randomUUID() + suffix
}

func randomUUID() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(crand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (c *Client) generateAndroidDeviceID() string {
	s := fmt.Sprintf("%d", time.Now().UnixNano())
	h := sha256.Sum256([]byte(s))
	return "android-" + fmt.Sprintf("%x", h)[:16]
}

func (c *Client) userID() int64 {
	if c.httpPrivate.Jar == nil {
		return 0
	}
	u, _ := url.Parse("https://" + igcfg.APIDomain)
	for _, ck := range c.httpPrivate.Jar.Cookies(u) {
		if ck.Name == "ds_user_id" {
			id, _ := strconv.ParseInt(ck.Value, 10, 64)
			return id
		}
	}
	if c.authData != nil {
		if s, ok := c.authData["ds_user_id"]; ok {
			id, _ := strconv.ParseInt(s, 10, 64)
			return id
		}
	}
	return 0
}

func (c *Client) sessionID() string {
	u, _ := url.Parse("https://" + igcfg.APIDomain)
	for _, ck := range c.httpPrivate.Jar.Cookies(u) {
		if ck.Name == "sessionid" {
			return ck.Value
		}
	}
	if c.authData != nil {
		return c.authData["sessionid"]
	}
	return ""
}

func (c *Client) loginSessionLooksValid() bool {
	if c.sessionID() != "" {
		return true
	}
	if c.userID() != 0 {
		return true
	}
	if len(c.authData) > 0 {
		if s := c.authData["sessionid"]; s != "" {
			return true
		}
		if s := c.authData["ds_user_id"]; s != "" {
			return true
		}
	}
	return false
}

func (c *Client) csrfToken() string {
	if c.token != "" {
		return c.token
	}
	u, _ := url.Parse("https://" + igcfg.APIDomain)
	for _, ck := range c.httpPrivate.Jar.Cookies(u) {
		if ck.Name == "csrftoken" {
			c.token = ck.Value
			return c.token
		}
	}
	c.token = igenc.GenToken(64)
	return c.token
}

func (c *Client) authorizationHeader() string {
	if len(c.authData) == 0 {
		return ""
	}
	j, err := json.Marshal(c.authData)
	if err != nil {
		return ""
	}
	b64 := base64.StdEncoding.EncodeToString(j)
	return "Bearer IGT:2:" + b64
}

// ParseAuthorizationHeader decodes a Bearer IGT:2:… Instagram authorization header into key-value pairs.
func ParseAuthorizationHeader(header string) map[string]string {
	if header == "" {
		return nil
	}
	h := strings.TrimSpace(header)
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		h = strings.TrimSpace(h[7:])
	}
	idx := strings.LastIndex(h, ":")
	if idx < 0 {
		return nil
	}
	b64 := h[idx+1:]
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil
	}
	var m map[string]string
	if json.Unmarshal(raw, &m) != nil {
		return nil
	}
	return m
}

func (c *Client) injectSessionIDToPublic() bool {
	sid := c.sessionID()
	if sid == "" {
		return false
	}
	u, _ := url.Parse(igcfg.PublicWebURL)
	c.httpPublic.Jar.SetCookies(u, []*http.Cookie{{Name: "sessionid", Value: sid}})
	return true
}

func (c *Client) randomDelay() {
	if c.delayMin > 0 && c.delayMax >= c.delayMin {
		d := c.delayMin + time.Duration(c.rng.Float64()*float64(c.delayMax-c.delayMin))
		time.Sleep(d)
	}
}

// PrivateRequestOpts configures a single private Instagram API call (POST or GET).
type PrivateRequestOpts struct {
	// Endpoint is the path after /api, with or without a leading slash (e.g. "feed/timeline/" or "/v1/feed/timeline/").
	Endpoint string
	// Data is JSON-encoded form body for POST when RawBody is empty.
	Data any
	// Params are query parameters; do not embed them in Endpoint.
	Params url.Values
	// Login skips some delays reserved for authenticated traffic.
	Login bool
	// WithSignature builds a signed_body form payload from Data when Data is a map.
	WithSignature bool
	// ExtraHeaders are merged into the request after base headers.
	ExtraHeaders http.Header
	// ExtraSig appends additional signed_body fragments after the signature block.
	ExtraSig []string
	// Domain overrides the API host (default i.instagram.com).
	Domain string
	// RawBody, when set, is sent as the POST body instead of Data.
	RawBody string
	// IsRawJSONBody treats Data as raw JSON string or marshals non-string values for the body.
	IsRawJSONBody bool
}

// PrivateRequest performs an authenticated request to the private API and returns the top-level JSON object.
func (c *Client) PrivateRequest(opts PrivateRequestOpts) (map[string]any, error) {
	var retried408 bool
	for {
		if !retried408 {
			c.randomDelay()
			if !opts.Login {
				time.Sleep(c.requestTimeout)
			}
		}
		endpoint := opts.Endpoint
		if !strings.HasPrefix(endpoint, "/") {
			endpoint = "/v1/" + endpoint
		}
		if endpoint == "/challenge/" {
			endpoint = "/v1/challenge/"
		}
		domain := opts.Domain
		if domain == "" {
			domain = igcfg.APIDomain
		}
		apiURL := "https://" + domain + "/api" + endpoint
		if strings.Contains(apiURL, "?") && opts.Params != nil {
			return nil, fmt.Errorf("ig: use Params field instead of query in endpoint")
		}
		u, err := url.Parse(apiURL)
		if err != nil {
			return nil, err
		}
		if opts.Params != nil {
			u.RawQuery = opts.Params.Encode()
		}

		reqHeaders := c.buildBaseHeaders()
		if auth := c.authorizationHeader(); auth != "" {
			reqHeaders.Set("Authorization", auth)
		}
		for k, vals := range opts.ExtraHeaders {
			for _, v := range vals {
				reqHeaders.Add(k, v)
			}
		}

		var body io.Reader
		var method string
		if opts.Data != nil || opts.RawBody != "" {
			method = http.MethodPost
			var bodyStr string
			if opts.RawBody != "" {
				bodyStr = opts.RawBody
			} else if opts.IsRawJSONBody {
				if s, ok := opts.Data.(string); ok {
					bodyStr = s
				} else {
					b, err := json.Marshal(opts.Data)
					if err != nil {
						return nil, err
					}
					bodyStr = string(b)
				}
			} else if m, ok := opts.Data.(map[string]any); ok {
				if opts.WithSignature {
					j, err := igenc.Dumps(m)
					if err != nil {
						return nil, err
					}
					bodyStr = igenc.GenerateSignature(j)
					for _, ex := range opts.ExtraSig {
						bodyStr += "&" + ex
					}
				} else {
					b, err := json.Marshal(m)
					if err != nil {
						return nil, err
					}
					var buf bytes.Buffer
					if err := json.Compact(&buf, b); err != nil {
						return nil, err
					}
					bodyStr = buf.String()
				}
			} else {
				j, err := igenc.Dumps(opts.Data)
				if err != nil {
					return nil, err
				}
				if opts.WithSignature {
					bodyStr = igenc.GenerateSignature(j)
					for _, ex := range opts.ExtraSig {
						bodyStr += "&" + ex
					}
				} else {
					bodyStr = j
				}
			}
			reqHeaders.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
			body = strings.NewReader(bodyStr)
		} else {
			method = http.MethodGet
			reqHeaders.Del("Content-Type")
		}

		req, err := http.NewRequest(method, u.String(), body)
		if err != nil {
			return nil, err
		}
		req.Header = reqHeaders

		c.privateReqCount++
		resp, err := c.httpPrivate.Do(req)
		if err != nil {
			return nil, &igerrors.ClientConnection{ClientError: igerrors.ClientError{Message: err.Error()}}
		}
		c.LastHTTPResponse = resp
		rawBytes, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, &igerrors.ClientConnection{ClientError: igerrors.ClientError{Message: readErr.Error()}}
		}
		raw := string(rawBytes)

		if mid := resp.Header.Get("ig-set-x-mid"); mid != "" {
			c.mid = mid
		}
		if auth := resp.Header.Get("ig-set-authorization"); auth != "" {
			if m := ParseAuthorizationHeader(auth); m != nil {
				c.authData = m
			}
		}

		var last map[string]any
		if err := json.Unmarshal(rawBytes, &last); err != nil {
			last = nil
		}
		c.LastJSON = last

		if resp.StatusCode == http.StatusRequestTimeout && !retried408 {
			retried408 = true
			time.Sleep(60 * time.Second)
			continue
		}

		if resp.StatusCode >= 400 {
			if err := igerrors.MapPrivateHTTPError(opts.Endpoint, resp, rawBytes); err != nil {
				return nil, err
			}
			return nil, &igerrors.ClientError{Status: resp.StatusCode, RawBody: raw, Endpoint: opts.Endpoint}
		}

		if last == nil {
			return nil, &igerrors.ClientJSONDecode{ClientError: igerrors.ClientError{Message: "invalid json", RawBody: raw}}
		}
		c.LastJSON = last
		if err := igerrors.CheckStatusFail(last, raw, opts.Endpoint); err != nil {
			return nil, err
		}
		return last, nil
	}
}

func (c *Client) mergeSettings(s *Settings) {
	c.uuids = s.UUIDs
	c.mid = s.Mid
	if s.IgURur != nil {
		c.igURur = *s.IgURur
	}
	if s.IgWWWClaim != nil {
		c.igWWWClaim = *s.IgWWWClaim
	}
	c.authData = s.AuthorizationData
	c.deviceSettings = s.DeviceSettings
	c.userAgent = s.UserAgent
	c.country = s.Country
	c.countryCode = s.CountryCode
	c.locale = s.Locale
	c.timezoneOffset = s.TimezoneOffset
	c.lastLogin = s.LastLogin
	c.bloksVersioningID = s.DeviceSettings.BloksVersioningID
	c.pickApp(s.UUIDs.UUID)
	if s.UserAgent != "" && !c.OverrideAppVersion {
		c.refreshUserAgent(s.UserAgent)
	} else {
		c.refreshUserAgent("")
	}
	u, _ := url.Parse("https://" + igcfg.APIDomain)
	cookieMap := map[string]string{}
	for name, val := range s.Cookies {
		if val != "" {
			cookieMap[name] = val
		}
	}
	if s.Mid != "" {
		cookieMap["mid"] = s.Mid
	}
	if s.AuthorizationData != nil {
		for _, key := range []string{"sessionid", "ds_user_id", "csrftoken"} {
			v := s.AuthorizationData[key]
			if v == "" {
				continue
			}
			if key == "sessionid" {
				if dec, err := url.QueryUnescape(v); err == nil && dec != "" {
					v = dec
				}
			}
			cookieMap[key] = v
		}
	}
	var cookies []*http.Cookie
	for name, val := range cookieMap {
		cookies = append(cookies, &http.Cookie{Name: name, Value: val})
	}
	c.httpPrivate.Jar.SetCookies(u, cookies)
}

func (c *Client) exportSettings() *Settings {
	u, _ := url.Parse("https://" + igcfg.APIDomain)
	cookieMap := map[string]string{}
	for _, ck := range c.httpPrivate.Jar.Cookies(u) {
		cookieMap[ck.Name] = ck.Value
	}
	var igRur *string
	if c.igURur != "" {
		s := c.igURur
		igRur = &s
	}
	var igW *string
	if c.igWWWClaim != "" {
		s := c.igWWWClaim
		igW = &s
	}
	ad := map[string]string{}
	for k, v := range c.authData {
		ad[k] = v
	}
	return &Settings{
		UUIDs:             c.uuids,
		Mid:               c.mid,
		IgURur:            igRur,
		IgWWWClaim:        igW,
		AuthorizationData: ad,
		Cookies:           cookieMap,
		LastLogin:         c.lastLogin,
		DeviceSettings:    c.deviceSettings,
		UserAgent:         c.userAgent,
		Country:           c.country,
		CountryCode:       c.countryCode,
		Locale:            c.locale,
		TimezoneOffset:    c.timezoneOffset,
	}
}

// LoadSettings reads a JSON session from path and applies it to the client. When overrideAppVersion is true, OverrideAppVersion is set so app version may be upgraded from config.
func (c *Client) LoadSettings(path string, overrideAppVersion bool) error {
	s, err := LoadSettingsFromFile(path)
	if err != nil {
		return err
	}
	c.OverrideAppVersion = overrideAppVersion
	c.mergeSettings(s)
	return nil
}

// DumpSettings writes the current session and device state to path as JSON.
func (c *Client) DumpSettings(path string) error {
	return DumpSettingsToFile(path, c.exportSettings())
}
