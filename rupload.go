package ig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	igcfg "github.com/felipeinf/instago/config"
	"github.com/felipeinf/instago/igerrors"
	_ "golang.org/x/image/webp"
)

const ruploadRetryContext = `{"num_step_auto_retry":0,"num_reupload":0,"num_step_manual_retry":0}`

func (c *Client) ruploadURL(kind, uploadName string) string {
	kind = strings.TrimPrefix(kind, "/")
	uploadName = strings.TrimPrefix(uploadName, "/")
	return "https://" + igcfg.APIDomain + "/" + kind + "/" + uploadName
}

func (c *Client) doRupload(reqMethod, fullURL string, body []byte, extra http.Header) (int, []byte, error) {
	c.randomDelay()
	time.Sleep(c.requestTimeout)
	h := c.buildBaseHeaders()
	if auth := c.authorizationHeader(); auth != "" {
		h.Set("Authorization", auth)
	}
	for k, vals := range extra {
		for _, v := range vals {
			h.Add(k, v)
		}
	}
	var rdr io.Reader
	if len(body) > 0 {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequest(reqMethod, fullURL, rdr)
	if err != nil {
		return 0, nil, err
	}
	req.Header = h
	resp, err := c.httpPrivate.Do(req)
	if err != nil {
		return 0, nil, &igerrors.ClientConnection{ClientError: igerrors.ClientError{Message: err.Error()}}
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return resp.StatusCode, nil, readErr
	}
	if resp.StatusCode >= 400 {
		return resp.StatusCode, raw, &igerrors.ClientError{Status: resp.StatusCode, RawBody: string(raw), Endpoint: fullURL}
	}
	return resp.StatusCode, raw, nil
}

func (c *Client) igPhotoRuploadName(uploadID string) string {
	suffix := c.rng.Intn(9000000000) + 1000000000
	return uploadID + "_0_" + strconv.Itoa(suffix)
}

// DirectPhotoRupload uploads image bytes for direct (or general photo pipeline). uploadID should be a numeric string (e.g. Unix ms). Returns width and height from the decoded image.
func (c *Client) DirectPhotoRupload(uploadID string, imageBytes []byte, entityType string) (width, height int, err error) {
	if c.userID() == 0 {
		return 0, 0, fmt.Errorf("ig: login required")
	}
	if uploadID == "" {
		uploadID = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	w, h, err := imageBoundsFromBytes(imageBytes)
	if err != nil {
		return 0, 0, err
	}
	if entityType == "" {
		entityType = "image/jpeg"
	}
	params := map[string]any{
		"retry_context":       ruploadRetryContext,
		"media_type":          "1",
		"xsharing_user_ids":   "[]",
		"upload_id":           uploadID,
		"image_compression":   `{"lib_name":"moz","lib_version":"3.1.m","quality":"80"}`,
	}
	pj, err := json.Marshal(params)
	if err != nil {
		return 0, 0, err
	}
	name := c.igPhotoRuploadName(uploadID)
	full := c.ruploadURL("rupload_igphoto", name)
	extra := http.Header{}
	extra.Set("Accept-Encoding", "gzip")
	extra.Set("X-Instagram-Rupload-Params", string(pj))
	extra.Set("X_FB_PHOTO_WATERFALL_ID", c.generateUUID("", ""))
	extra.Set("X-Entity-Type", entityType)
	extra.Set("Offset", "0")
	extra.Set("X-Entity-Name", name)
	extra.Set("X-Entity-Length", strconv.Itoa(len(imageBytes)))
	extra.Set("Content-Type", "application/octet-stream")
	extra.Set("Content-Length", strconv.Itoa(len(imageBytes)))
	code, _, err := c.doRupload(http.MethodPost, full, imageBytes, extra)
	if err != nil {
		return 0, 0, err
	}
	if code != http.StatusOK {
		return 0, 0, fmt.Errorf("ig: photo rupload http %d", code)
	}
	return w, h, nil
}

// DirectVideoRupload uploads a direct video (MP4). uploadID should be a numeric string. meta must carry width, height, and duration.
func (c *Client) DirectVideoRupload(uploadID string, videoBytes []byte, meta VideoUploadMeta) error {
	if c.userID() == 0 {
		return fmt.Errorf("ig: login required")
	}
	if uploadID == "" {
		uploadID = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}
	if meta.WidthPx <= 0 || meta.HeightPx <= 0 || meta.DurationMS <= 0 {
		return fmt.Errorf("ig: VideoUploadMeta width, height, duration required for video rupload")
	}
	uid := c.userID()
	xshare, _ := json.Marshal([]int64{uid})
	params := map[string]any{
		"retry_context":            ruploadRetryContext,
		"media_type":               "2",
		"xsharing_user_ids":        string(xshare),
		"upload_id":                uploadID,
		"upload_media_duration_ms": strconv.Itoa(meta.DurationMS),
		"upload_media_width":       strconv.Itoa(meta.WidthPx),
		"upload_media_height":      strconv.Itoa(meta.HeightPx),
		"direct_v2":                "1",
	}
	pj, err := json.Marshal(params)
	if err != nil {
		return err
	}
	name := c.igPhotoRuploadName(uploadID)
	base := c.ruploadURL("rupload_igvideo", name)
	extra := http.Header{}
	extra.Set("Accept-Encoding", "gzip, deflate")
	extra.Set("X-Instagram-Rupload-Params", string(pj))
	extra.Set("X_FB_VIDEO_WATERFALL_ID", c.generateUUID("", ""))
	code, _, err := c.doRupload(http.MethodGet, base, nil, extra)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("ig: video rupload init http %d", code)
	}
	vlen := len(videoBytes)
	postExtra := http.Header{}
	for k, vals := range extra {
		for _, v := range vals {
			postExtra.Add(k, v)
		}
	}
	postExtra.Set("Offset", "0")
	postExtra.Set("X-Entity-Name", name)
	postExtra.Set("X-Entity-Length", strconv.Itoa(vlen))
	postExtra.Set("Content-Type", "application/octet-stream")
	postExtra.Set("Content-Length", strconv.Itoa(vlen))
	postExtra.Set("X-Entity-Type", "video/mp4")
	code2, _, err := c.doRupload(http.MethodPost, base, videoBytes, postExtra)
	if err != nil {
		return err
	}
	if code2 != http.StatusOK {
		return fmt.Errorf("ig: video rupload post http %d", code2)
	}
	return nil
}

func imageBoundsFromBytes(b []byte) (w, h int, err error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(b))
	if err != nil {
		return 0, 0, fmt.Errorf("ig: decode image config: %w", err)
	}
	return cfg.Width, cfg.Height, nil
}

func directPhotoEntityType(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg", nil
	case ".png":
		return "image/png", nil
	case ".webp":
		return "image/webp", nil
	default:
		return "", fmt.Errorf("ig: unsupported image extension %q (use jpg, png, webp)", ext)
	}
}
