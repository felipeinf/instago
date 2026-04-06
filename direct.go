package ig

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	defaultDirectInboxLimit          = 20
	defaultDirectThreadPreviewLimit  = 10
	defaultDirectThreadMessagesLimit = 20
)

var directHTTPURL = regexp.MustCompile(`https?://[^\s]+`)

// DirectInboxOptions configures GET direct_v2/inbox/. Zero values use Instagram-style defaults for a single request (up to 20 threads, 10 preview messages per thread).
type DirectInboxOptions struct {
	SelectedFilter     string
	Box                string
	ThreadMessageLimit int
	Limit              int
	Cursor             string
}

// DirectInboxChunk returns one page of DM threads. Defaults: limit=20, thread_message_limit=10.
func (c *Client) DirectInboxChunk(opt DirectInboxOptions) ([]DirectThread, string, error) {
	if c.userID() == 0 {
		return nil, "", fmt.Errorf("ig: login required for direct inbox")
	}
	limit := opt.Limit
	if limit <= 0 {
		limit = defaultDirectInboxLimit
	}
	tml := opt.ThreadMessageLimit
	if tml <= 0 {
		tml = defaultDirectThreadPreviewLimit
	}
	params := url.Values{}
	params.Set("visual_message_return_type", "unseen")
	params.Set("thread_message_limit", strconv.Itoa(tml))
	params.Set("persistentBadging", "true")
	params.Set("limit", strconv.Itoa(limit))
	params.Set("is_prefetching", "false")
	if opt.SelectedFilter != "" {
		params.Set("selected_filter", opt.SelectedFilter)
	}
	if opt.Box == "general" {
		params.Set("folder", "1")
	} else if opt.Box == "primary" {
		params.Set("folder", "0")
	}
	if opt.Cursor != "" {
		params.Set("cursor", opt.Cursor)
		params.Set("direction", "older")
		params.Set("fetch_reason", "page_scroll")
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "direct_v2/inbox/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, "", err
	}
	inbox, _ := res["inbox"].(map[string]any)
	var threads []DirectThread
	if inbox != nil {
		for _, t := range toSliceMap("threads", inbox) {
			threads = append(threads, extractDirectThreadMap(t))
		}
	}
	next := ""
	if inbox != nil {
		next = toString(inbox["oldest_cursor"])
	}
	return threads, next, nil
}

// DirectPendingChunk returns one page of pending DM requests.
func (c *Client) DirectPendingChunk(cursor string) ([]DirectThread, string, error) {
	if c.userID() == 0 {
		return nil, "", fmt.Errorf("ig: login required")
	}
	params := url.Values{}
	params.Set("visual_message_return_type", "unseen")
	params.Set("persistentBadging", "true")
	params.Set("is_prefetching", "false")
	params.Set("request_session_id", c.uuids.RequestID)
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "direct_v2/pending_inbox/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, "", err
	}
	inbox, _ := res["inbox"].(map[string]any)
	var threads []DirectThread
	if inbox != nil {
		for _, t := range toSliceMap("threads", inbox) {
			threads = append(threads, extractDirectThreadMap(t))
		}
	}
	next := ""
	if inbox != nil {
		next = toString(inbox["oldest_cursor"])
	}
	return threads, next, nil
}

// DirectSpamChunk returns one page of hidden/spam DM threads.
func (c *Client) DirectSpamChunk(cursor string) ([]DirectThread, string, error) {
	if c.userID() == 0 {
		return nil, "", fmt.Errorf("ig: login required")
	}
	params := url.Values{}
	params.Set("visual_message_return_type", "unseen")
	params.Set("persistentBadging", "true")
	params.Set("is_prefetching", "false")
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "direct_v2/spam_inbox/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, "", err
	}
	inbox, _ := res["inbox"].(map[string]any)
	var threads []DirectThread
	if inbox != nil {
		for _, t := range toSliceMap("threads", inbox) {
			threads = append(threads, extractDirectThreadMap(t))
		}
	}
	next := ""
	if inbox != nil {
		next = toString(inbox["oldest_cursor"])
	}
	return threads, next, nil
}

// DirectThreadOptions configures GET direct_v2/threads/{id}/. Limit defaults to 20 (one typical app request; not a fixed message count).
type DirectThreadOptions struct {
	Cursor string
	Limit  int
}

// DirectThreadPage fetches one page of messages for a thread. With zero options this performs a single request (limit=20).
func (c *Client) DirectThreadPage(threadID int64, opt DirectThreadOptions) (DirectThread, string, error) {
	if c.userID() == 0 {
		return DirectThread{}, "", fmt.Errorf("ig: login required")
	}
	limit := opt.Limit
	if limit <= 0 {
		limit = defaultDirectThreadMessagesLimit
	}
	params := url.Values{}
	params.Set("visual_message_return_type", "unseen")
	params.Set("direction", "older")
	params.Set("seq_id", "40065")
	params.Set("limit", strconv.Itoa(limit))
	if opt.Cursor != "" {
		params.Set("cursor", opt.Cursor)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("direct_v2/threads/%d/", threadID),
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return DirectThread{}, "", err
	}
	th, _ := res["thread"].(map[string]any)
	if th == nil {
		return DirectThread{}, "", fmt.Errorf("ig: direct thread missing thread object")
	}
	dt := extractDirectThreadMap(th)
	next := toString(th["oldest_cursor"])
	return dt, next, nil
}

// DirectSendText sends a DM. Pass either userIDs (new thread) or threadIDs (existing), not both.
func (c *Client) DirectSendText(text string, userIDs []int64, threadIDs []int64) (DirectMessage, error) {
	if c.userID() == 0 {
		return DirectMessage{}, fmt.Errorf("ig: login required")
	}
	hasU, hasT := len(userIDs) > 0, len(threadIDs) > 0
	if hasU == hasT {
		return DirectMessage{}, fmt.Errorf("ig: specify exactly one of userIDs or threadIDs")
	}
	token := c.generateUUID("", "")
	method := "text"
	kw := map[string]any{
		"action":               "send_item",
		"is_x_transport_forward": "false",
		"send_silently":        "false",
		"is_shh_mode":          "0",
		"send_attribution":     "message_button",
		"client_context":       token,
		"device_id":            c.uuids.AndroidDeviceID,
		"mutation_token":       token,
		"btt_dual_send":        "false",
		"nav_chain":            "1qT:feed_timeline:1,1qT:feed_timeline:2,1qT:feed_timeline:3,7Az:direct_inbox:4,7Az:direct_inbox:5,5rG:direct_thread:7",
		"is_ae_dual_send":      "false",
		"offline_threading_id":   token,
	}
	if directHTTPURL.MatchString(text) {
		method = "link"
		urls := directHTTPURL.FindAllString(text, -1)
		ub, _ := json.Marshal(urls)
		kw["link_text"] = text
		kw["link_urls"] = string(ub)
	} else {
		kw["text"] = text
	}
	if len(threadIDs) > 0 {
		tb, _ := json.Marshal(threadIDs)
		kw["thread_ids"] = string(tb)
	}
	if len(userIDs) > 0 {
		inner := make([][]int64, 1)
		inner[0] = userIDs
		rb, _ := json.Marshal(inner)
		kw["recipient_users"] = string(rb)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("direct_v2/threads/broadcast/%s/", method),
		Data:          c.withDefaultData(kw),
		WithSignature: false,
	})
	if err != nil {
		return DirectMessage{}, err
	}
	payload, _ := res["payload"].(map[string]any)
	if payload == nil {
		return DirectMessage{}, fmt.Errorf("ig: direct send missing payload")
	}
	tid := ""
	if len(threadIDs) == 1 {
		tid = strconv.FormatInt(threadIDs[0], 10)
	}
	return extractDirectMessage(payload, tid), nil
}

var directNavChainsPhoto = []string{
	"6xQ:direct_media_picker_photos_fragment:1,5rG:direct_thread:2,5ME:direct_quick_camera_fragment:3,5ME:direct_quick_camera_fragment:4,4ju:reel_composer_preview:5,5rG:direct_thread:6,5rG:direct_thread:7,6xQ:direct_media_picker_photos_fragment:8,5rG:direct_thread:9",
	"1qT:feed_timeline:1,7Az:direct_inbox:2,7Az:direct_inbox:3,5rG:direct_thread:4,6xQ:direct_media_picker_photos_fragment:5,5rG:direct_thread:6,5rG:direct_thread:7,6xQ:direct_media_picker_photos_fragment:8,5rG:direct_thread:9",
}

// DirectSendPhoto uploads a local image (JPEG, PNG, or WebP) and sends it in a DM. Specify either userIDs or threadIDs, not both.
func (c *Client) DirectSendPhoto(path string, userIDs []int64, threadIDs []int64) (DirectMessage, error) {
	if c.userID() == 0 {
		return DirectMessage{}, fmt.Errorf("ig: login required")
	}
	hasU, hasT := len(userIDs) > 0, len(threadIDs) > 0
	if hasU == hasT {
		return DirectMessage{}, fmt.Errorf("ig: specify exactly one of userIDs or threadIDs")
	}
	entityType, err := directPhotoEntityType(path)
	if err != nil {
		return DirectMessage{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return DirectMessage{}, err
	}
	uploadID := strconv.FormatInt(time.Now().UnixMilli(), 10)
	if _, _, err := c.DirectPhotoRupload(uploadID, raw, entityType); err != nil {
		return DirectMessage{}, err
	}
	token := c.generateUUID("", "")
	nav := directNavChainsPhoto[c.rng.Intn(len(directNavChainsPhoto))]
	data := map[string]any{
		"action":                  "send_item",
		"is_shh_mode":             "0",
		"send_attribution":        "inbox",
		"client_context":          token,
		"mutation_token":          token,
		"nav_chain":               nav,
		"offline_threading_id":    token,
		"allow_full_aspect_ratio": "true",
		"upload_id":               uploadID,
	}
	if len(threadIDs) > 0 {
		tb, _ := json.Marshal(threadIDs)
		data["thread_ids"] = string(tb)
	}
	if len(userIDs) > 0 {
		inner := [][]int64{userIDs}
		rb, _ := json.Marshal(inner)
		data["recipient_users"] = string(rb)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "direct_v2/threads/broadcast/configure_photo/",
		Data:          c.withDefaultData(data),
		WithSignature: false,
	})
	if err != nil {
		return DirectMessage{}, err
	}
	payload, _ := res["payload"].(map[string]any)
	if payload == nil {
		return DirectMessage{}, fmt.Errorf("ig: direct photo missing payload")
	}
	tid := ""
	if len(threadIDs) == 1 {
		tid = strconv.FormatInt(threadIDs[0], 10)
	}
	return extractDirectMessage(payload, tid), nil
}

// DirectSendVideo uploads a local MP4 and sends it in a DM. meta must include width, height, and duration in milliseconds (probe externally if needed). Specify either userIDs or threadIDs, not both.
func (c *Client) DirectSendVideo(path string, meta VideoUploadMeta, userIDs []int64, threadIDs []int64) (DirectMessage, error) {
	if c.userID() == 0 {
		return DirectMessage{}, fmt.Errorf("ig: login required")
	}
	hasU, hasT := len(userIDs) > 0, len(threadIDs) > 0
	if hasU == hasT {
		return DirectMessage{}, fmt.Errorf("ig: specify exactly one of userIDs or threadIDs")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return DirectMessage{}, err
	}
	uploadID := strconv.FormatInt(time.Now().UnixMilli(), 10)
	if err := c.DirectVideoRupload(uploadID, raw, meta); err != nil {
		return DirectMessage{}, err
	}
	token := c.generateUUID("", "")
	nav := directNavChainsPhoto[c.rng.Intn(len(directNavChainsPhoto))]
	data := map[string]any{
		"action":               "send_item",
		"is_shh_mode":          "0",
		"send_attribution":     "direct_thread",
		"client_context":       token,
		"mutation_token":       token,
		"nav_chain":            nav,
		"offline_threading_id": token,
		"video_result":         "",
		"upload_id":            uploadID,
	}
	if len(threadIDs) > 0 {
		tb, _ := json.Marshal(threadIDs)
		data["thread_ids"] = string(tb)
	}
	if len(userIDs) > 0 {
		inner := [][]int64{userIDs}
		rb, _ := json.Marshal(inner)
		data["recipient_users"] = string(rb)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "direct_v2/threads/broadcast/configure_video/",
		Data:          c.withDefaultData(data),
		WithSignature: false,
	})
	if err != nil {
		return DirectMessage{}, err
	}
	payload, _ := res["payload"].(map[string]any)
	if payload == nil {
		return DirectMessage{}, fmt.Errorf("ig: direct video missing payload")
	}
	tid := ""
	if len(threadIDs) == 1 {
		tid = strconv.FormatInt(threadIDs[0], 10)
	}
	return extractDirectMessage(payload, tid), nil
}

// DirectMessageMediaURL returns the best URL to download image or video from a received DM (expires; use session cookies if required).
func DirectMessageMediaURL(m DirectMessage) string {
	if m.VisualVideoURL != "" {
		return m.VisualVideoURL
	}
	if m.VisualPhotoURL != "" {
		return m.VisualPhotoURL
	}
	if m.MediaVideoURL != "" {
		return m.MediaVideoURL
	}
	return m.MediaURL
}
