package password

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// RSADecodePublicKeyFromBase64 parses a base64-encoded PEM PKIX public key as returned by Instagram headers.
func RSADecodePublicKeyFromBase64(pubKeyBase64 string) (*rsa.PublicKey, error) {
	pubKey, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubKey)
	if block == nil {
		return nil, fmt.Errorf("password: invalid PEM in public key")
	}
	pKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return pKey.(*rsa.PublicKey), nil
}

func aesGCMEncrypt(key, plaintext, additionalData []byte) (iv, ciphertext, tag []byte, err error) {
	iv = make([]byte, 12)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, err
	}
	sealed := gcm.Seal(nil, iv, plaintext, additionalData)
	if len(sealed) < 16 {
		return nil, nil, nil, fmt.Errorf("password: sealed too short")
	}
	tag = sealed[len(sealed)-16:]
	ciphertext = sealed[:len(sealed)-16]
	return iv, ciphertext, tag, nil
}

func rsaEncryptPKCS1v15(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

// EncryptPassword returns the #PWD_INSTAGRAM:4:… enc_password value for login; ts defaults to Unix seconds if empty.
func EncryptPassword(password, pubKeyEncoded string, pubKeyVersion int, ts string) (string, error) {
	if ts == "" {
		ts = strconv.FormatInt(time.Now().Unix(), 10)
	}
	publicKey, err := RSADecodePublicKeyFromBase64(pubKeyEncoded)
	if err != nil {
		return "", err
	}
	sessionKey := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, sessionKey); err != nil {
		return "", err
	}
	rsaEncrypted, err := rsaEncryptPKCS1v15(publicKey, sessionKey)
	if err != nil {
		return "", err
	}
	sizeBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(sizeBuf, uint16(len(rsaEncrypted)))
	iv, enc, tag, err := aesGCMEncrypt(sessionKey, []byte(password), []byte(ts))
	if err != nil {
		return "", err
	}
	payload := []byte{1, byte(pubKeyVersion)}
	payload = append(payload, iv...)
	payload = append(payload, sizeBuf...)
	payload = append(payload, rsaEncrypted...)
	payload = append(payload, tag...)
	payload = append(payload, enc...)
	encoded := base64.StdEncoding.EncodeToString(payload)
	return fmt.Sprintf("#PWD_INSTAGRAM:4:%s:%s", ts, encoded), nil
}

// FetchPublicKeys performs a GET to qe/sync/ and reads ig-set-password-encryption-key-id and ig-set-password-encryption-pub-key.
func FetchPublicKeys(client *http.Client) (keyID int, pubKey string, err error) {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequest(http.MethodGet, "https://i.instagram.com/api/v1/qe/sync/", nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	idStr := resp.Header.Get("ig-set-password-encryption-key-id")
	pub := resp.Header.Get("ig-set-password-encryption-pub-key")
	if idStr == "" || pub == "" {
		return 0, "", fmt.Errorf("password: missing encryption headers")
	}
	keyID, err = strconv.Atoi(idStr)
	if err != nil {
		return 0, "", err
	}
	return keyID, pub, nil
}
