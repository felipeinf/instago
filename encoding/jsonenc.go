package encoding

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
)

// Dumps JSON-encodes v with HTML escaping disabled, trims a trailing newline, and returns compact JSON text.
func Dumps(v any) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	s := buf.String()
	if len(s) > 0 && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	compact, err := compactJSON(s)
	if err != nil {
		return "", err
	}
	return string(compact), nil
}

func compactJSON(s string) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateSignature returns the signed_body form field value for Instagram private POST bodies (placeholder signature + URL-encoded JSON).
func GenerateSignature(data string) string {
	return "signed_body=SIGNATURE." + url.QueryEscape(data)
}

// GenerateJazoest builds the jazoest field from the sum of rune codes in symbols (Instagram login format).
func GenerateJazoest(symbols string) string {
	sum := 0
	for _, r := range symbols {
		sum += int(r)
	}
	return "2" + strconv.Itoa(sum)
}

// GenToken returns a random alphanumeric string of length size (cryptographic where possible).
func GenToken(size int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	max := big.NewInt(int64(len(chars)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			b[i] = chars[i%len(chars)]
			continue
		}
		b[i] = chars[n.Int64()]
	}
	return string(b)
}

const instagramIDAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"

// InstagramIDEncode converts a numeric media/user pk to the shortcode alphabet used in URLs.
func InstagramIDEncode(num int64) string {
	if num == 0 {
		return string(instagramIDAlphabet[0])
	}
	var arr []byte
	base := int64(len(instagramIDAlphabet))
	n := num
	for n > 0 {
		rem := n % base
		n /= base
		arr = append(arr, instagramIDAlphabet[rem])
	}
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return string(arr)
}

// InstagramIDDecode parses a shortcode string back to int64.
func InstagramIDDecode(shortcode string) (int64, error) {
	base := int64(len(instagramIDAlphabet))
	var num int64
	for _, char := range shortcode {
		idx := bytes.IndexRune([]byte(instagramIDAlphabet), char)
		if idx < 0 {
			return 0, fmt.Errorf("invalid char %q", char)
		}
		num = num*base + int64(idx)
	}
	return num, nil
}
