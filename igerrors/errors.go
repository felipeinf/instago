package igerrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ClientError is the base error for Instagram HTTP failures with optional parsed JSON body.
type ClientError struct {
	Status   int
	Message  string
	Body     map[string]any
	RawBody  string
	Endpoint string
}

func (e *ClientError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("instagram error: status %d", e.Status)
}

// TwoFactorRequired indicates the account needs a second factor; TwoFactorIdentifier may be set from the response.
type TwoFactorRequired struct {
	ClientError
	TwoFactorIdentifier string
}

// ChallengeRequired indicates Instagram returned challenge_required in the message.
type ChallengeRequired struct {
	ClientError
}

// LoginRequired means the session is invalid or absent for the requested resource.
type LoginRequired struct {
	ClientError
}

// BadPassword is returned when error_type is bad_password.
type BadPassword struct {
	ClientError
}

// UserNotFound means the user does not exist or cannot be loaded.
type UserNotFound struct {
	ClientError
	Username string
	UserID   string
}

// PrivateAccount means the viewer cannot access a private profile's data.
type PrivateAccount struct {
	ClientError
}

// RateLimitError is a structured rate limit response from the API.
type RateLimitError struct {
	ClientError
}

// PleaseWaitFewMinutes is returned when the message asks to wait before retrying.
type PleaseWaitFewMinutes struct {
	ClientError
}

// ClientThrottled wraps HTTP 429 responses.
type ClientThrottled struct {
	ClientError
}

// ClientJSONDecode means the response body was not valid JSON where JSON was expected.
type ClientJSONDecode struct {
	ClientError
}

func (e *ClientJSONDecode) Error() string {
	msg := e.Message
	if msg == "" {
		msg = "invalid json"
	}
	if e.RawBody == "" {
		return msg
	}
	prefix := e.RawBody
	const max = 400
	runes := []rune(prefix)
	if len(runes) > max {
		prefix = string(runes[:max]) + "..."
	}
	return msg + " (body): " + prefix
}

// ClientConnection wraps a transport-level failure message.
type ClientConnection struct {
	ClientError
}

// ClientRequestTimeout wraps HTTP 408.
type ClientRequestTimeout struct {
	ClientError
}

// BadCredentials is a local validation error (e.g. missing username).
type BadCredentials struct {
	Msg string
}

func (e *BadCredentials) Error() string {
	return e.Msg
}

// MapPrivateHTTPError maps HTTP status and JSON body to a concrete error type when recognized.
func MapPrivateHTTPError(endpoint string, resp *http.Response, body []byte) error {
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	raw := string(body)
	var last map[string]any
	_ = json.Unmarshal(body, &last)
	msg := ""
	if last != nil {
		if m, ok := last["message"].(string); ok {
			msg = m
		}
	}
	base := ClientError{
		Status:   status,
		Message:  msg,
		Body:     last,
		RawBody:  raw,
		Endpoint: endpoint,
	}
	switch status {
	case http.StatusUnauthorized:
		return &LoginRequired{ClientError: base}
	case http.StatusBadRequest:
		if strings.Contains(msg, "Please wait a few minutes") {
			return &PleaseWaitFewMinutes{ClientError: base}
		}
		if tf, ok := last["two_factor_info"].(map[string]any); ok {
			id, _ := tf["two_factor_identifier"].(string)
			return &TwoFactorRequired{ClientError: base, TwoFactorIdentifier: id}
		}
		if last != nil {
			et, _ := last["error_type"].(string)
			if et == "two_factor_required" {
				return &TwoFactorRequired{ClientError: base}
			}
			if et == "bad_password" {
				return &BadPassword{ClientError: base}
			}
			if et == "rate_limit_error" {
				return &RateLimitError{ClientError: base}
			}
		}
		if msg == "challenge_required" {
			return &ChallengeRequired{ClientError: base}
		}
		if strings.Contains(msg, "Not authorized to view user") {
			return &PrivateAccount{ClientError: base}
		}
		if strings.Contains(msg, "unable to fetch followers") {
			return &UserNotFound{ClientError: base}
		}
		return &ClientError{Status: status, Message: msg, Body: last, RawBody: raw, Endpoint: endpoint}
	case http.StatusForbidden:
		if msg == "login_required" {
			return &LoginRequired{ClientError: base}
		}
		return &ClientError{Status: status, Message: msg, Body: last, RawBody: raw, Endpoint: endpoint}
	case http.StatusTooManyRequests:
		return &ClientThrottled{ClientError: base}
	case http.StatusRequestTimeout:
		return &ClientRequestTimeout{ClientError: base}
	case http.StatusNotFound:
		return &UserNotFound{ClientError: base}
	default:
		if status >= 400 {
			return &ClientError{Status: status, Message: msg, Body: last, RawBody: raw, Endpoint: endpoint}
		}
	}
	return nil
}

// IsNotFound reports whether err is HTTP 404 as ClientError or a *UserNotFound.
func IsNotFound(err error) bool {
	var ce *ClientError
	if errors.As(err, &ce) && ce.Status == http.StatusNotFound {
		return true
	}
	var u *UserNotFound
	return errors.As(err, &u)
}

// AsUserNotFound sets *target when err unwraps to *UserNotFound.
func AsUserNotFound(err error, target **UserNotFound) bool {
	var u *UserNotFound
	if errors.As(err, &u) {
		*target = u
		return true
	}
	return false
}

// CheckStatusFail returns ClientError when JSON has status "fail" or includes error_title (HTTP 200 error payloads).
func CheckStatusFail(last map[string]any, raw string, endpoint string) error {
	if last == nil {
		return nil
	}
	if st, ok := last["status"].(string); ok && st == "fail" {
		return &ClientError{
			Status:   200,
			Message:  fmt.Sprint(last["message"]),
			Body:     last,
			RawBody:  raw,
			Endpoint: endpoint,
		}
	}
	if _, ok := last["error_title"]; ok {
		return &ClientError{
			Status:   200,
			Message:  fmt.Sprint(last["message"]),
			Body:     last,
			RawBody:  raw,
			Endpoint: endpoint,
		}
	}
	return nil
}
