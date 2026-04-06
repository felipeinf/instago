// Package igerrors defines typed errors returned from Instagram HTTP and JSON responses.
//
// Use errors.As to inspect specific failures (for example TwoFactorRequired, LoginRequired, UserNotFound).
// MapPrivateHTTPError maps HTTP status and body to these types where applicable.
package igerrors
