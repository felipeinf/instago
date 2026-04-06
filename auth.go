package ig

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	igcfg "github.com/felipeinf/instago/config"
	igenc "github.com/felipeinf/instago/encoding"
	"github.com/felipeinf/instago/igerrors"
	"github.com/felipeinf/instago/password"
)

// PreLoginFlow runs launcher/sync before login; throttling errors are ignored so login can proceed.
func (c *Client) PreLoginFlow() error {
	_, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint: "launcher/sync/",
		Data: map[string]any{
			"id":                      c.uuids.UUID,
			"server_config_retrieval": "1",
		},
		Login:         true,
		WithSignature: true,
	})
	if err != nil {
		var pw *igerrors.PleaseWaitFewMinutes
		var th *igerrors.ClientThrottled
		if errors.As(err, &pw) || errors.As(err, &th) {
			return nil
		}
		return err
	}
	return nil
}

// Login signs in with username and passwordPlain. If two-factor is required, call again with verificationCode non-empty.
func (c *Client) Login(username, passwordPlain, verificationCode string) error {
	if username != "" {
		c.username = username
	}
	if passwordPlain != "" {
		c.password = passwordPlain
	}
	if c.username == "" || c.password == "" {
		return &igerrors.BadCredentials{Msg: "Both username and password must be provided."}
	}
	if uid := c.userID(); uid != 0 {
		return nil
	}
	if err := c.PreLoginFlow(); err != nil {
		return err
	}
	keyID, pubKey, err := password.FetchPublicKeys(c.httpPublic)
	if err != nil {
		return err
	}
	encPass, err := password.EncryptPassword(c.password, pubKey, keyID, "")
	if err != nil {
		return err
	}
	data := map[string]any{
		"jazoest":             igenc.GenerateJazoest(c.uuids.PhoneID),
		"country_codes":       fmt.Sprintf(`[{"country_code":"%d","source":["default"]}]`, c.countryCode),
		"phone_id":            c.uuids.PhoneID,
		"enc_password":        encPass,
		"username":            c.username,
		"adid":                c.uuids.AdvertisingID,
		"guid":                c.uuids.UUID,
		"device_id":           c.uuids.AndroidDeviceID,
		"google_tokens":       "[]",
		"login_attempt_count": "0",
	}
	_, err = c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "accounts/login/",
		Data:          data,
		Login:         true,
		WithSignature: true,
	})
	if err != nil {
		var tf *igerrors.TwoFactorRequired
		if !errors.As(err, &tf) {
			return err
		}
		if strings.TrimSpace(verificationCode) == "" {
			return fmt.Errorf("%w (provide verification code)", err)
		}
		last := c.LastJSON
		tfInfo, _ := last["two_factor_info"].(map[string]any)
		tfID, _ := tfInfo["two_factor_identifier"].(string)
		twoData := map[string]any{
			"verification_code":     verificationCode,
			"phone_id":              c.uuids.PhoneID,
			"_csrftoken":            c.csrfToken(),
			"two_factor_identifier": tfID,
			"username":              c.username,
			"trust_this_device":     "0",
			"guid":                  c.uuids.UUID,
			"device_id":             c.uuids.AndroidDeviceID,
			"waterfall_id":          randomUUID(),
			"verification_method":   "3",
		}
		_, err = c.PrivateRequest(PrivateRequestOpts{
			Endpoint:      "accounts/two_factor_login/",
			Data:          twoData,
			Login:         true,
			WithSignature: true,
		})
		if err != nil {
			return err
		}
	}
	_ = c.LoginFlow()
	if !c.loginSessionLooksValid() {
		return fmt.Errorf("ig: login response missing session identity (ds_user_id / sessionid)")
	}
	now := float64(time.Now().Unix())
	c.lastLogin = &now
	return nil
}

// LoginFlow performs lightweight post-login feed calls (reels tray and timeline) to mimic the app.
func (c *Client) LoginFlow() error {
	_, _ = c.GetReelsTrayFeed("cold_start")
	_, _ = c.GetTimelineFeed("cold_start_fetch", "")
	return nil
}

// GetReelsTrayFeed calls feed/reels_tray/ with the given reason (e.g. "cold_start").
func (c *Client) GetReelsTrayFeed(reason string) (map[string]any, error) {
	var impressions any = map[string]any{}
	if reason != "cold_start" {
		impressions = map[string]any{strconv.FormatInt(c.userID(), 10): strconv.FormatInt(time.Now().Unix(), 10)}
	}
	data := map[string]any{
		"supported_capabilities_new": igcfg.SupportedCapabilities,
		"reason":                     reason,
		"timezone_offset":            strconv.Itoa(c.timezoneOffset),
		"tray_session_id":            c.uuids.TraySessionID,
		"request_id":                 c.uuids.RequestID,
		"page_size":                  50,
		"_uuid":                      c.uuids.UUID,
		"reel_tray_impressions":      impressions,
	}
	return c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "feed/reels_tray/",
		Data:          data,
		WithSignature: true,
	})
}

// GetTimelineFeed calls feed/timeline/. Pass maxID for pagination (reason becomes "pagination").
func (c *Client) GetTimelineFeed(reason, maxID string) (map[string]any, error) {
	h := http.Header{}
	h.Set("X-Ads-Opt-Out", "0")
	h.Set("X-DEVICE-ID", c.uuids.UUID)
	h.Set("X-CM-Bandwidth-KBPS", "-1.000")
	h.Set("X-CM-Latency", strconv.Itoa(c.rng.Intn(5)+1))
	data := map[string]any{
		"has_camera_permission": "1",
		"feed_view_info":        "[]",
		"phone_id":              c.uuids.PhoneID,
		"reason":                reason,
		"battery_level":         100,
		"timezone_offset":       strconv.Itoa(c.timezoneOffset),
		"device_id":             c.uuids.UUID,
		"request_id":            c.uuids.RequestID,
		"_uuid":                 c.uuids.UUID,
		"is_charging":           c.rng.Intn(2),
		"is_dark_mode":          1,
		"will_sound_on":         c.rng.Intn(2),
		"session_id":            c.uuids.ClientSessionID,
		"bloks_versioning_id":   c.bloksVersioningID,
	}
	if reason == "pull_to_refresh" || reason == "auto_refresh" {
		data["is_pull_to_refresh"] = "1"
	} else {
		data["is_pull_to_refresh"] = "0"
	}
	if maxID != "" {
		data["max_id"] = maxID
		data["reason"] = "pagination"
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "feed/timeline/",
		RawBody:       string(raw),
		WithSignature: false,
		ExtraHeaders:  h,
	})
}

// Logout calls accounts/logout/ and reports whether status is "ok".
func (c *Client) Logout() (bool, error) {
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "accounts/logout/",
		Data:          map[string]any{"one_tap_app_login": true},
		WithSignature: true,
	})
	if err != nil {
		return false, err
	}
	st, _ := res["status"].(string)
	return st == "ok", nil
}
