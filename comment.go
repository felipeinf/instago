package ig

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) resolveMediaRESTID(mediaID string) (string, error) {
	s := strings.TrimSpace(mediaID)
	if s == "" {
		return "", fmt.Errorf("ig: empty media id")
	}
	if strings.Contains(s, "_") {
		return s, nil
	}
	pk, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return "", fmt.Errorf("ig: media id: %w", err)
	}
	m, err := c.MediaInfoV1(pk)
	if err != nil {
		return "", err
	}
	if m.User.PK == 0 {
		return "", fmt.Errorf("ig: could not resolve media owner for comments")
	}
	return fmt.Sprintf("%d_%d", m.PK, m.User.PK), nil
}

func parseMediaCommentsResponse(res map[string]any) MediaCommentsPage {
	page := MediaCommentsPage{}
	if arr, ok := res["comments"].([]any); ok {
		for _, x := range arr {
			if m, ok := x.(map[string]any); ok {
				page.Comments = append(page.Comments, extractComment(m))
			}
		}
	}
	page.NextMaxID = toString(res["next_max_id"])
	page.NextMinID = toString(res["next_min_id"])
	page.CommentCount = int(toInt64(res["comment_count"]))
	moreMax := toBool(res["has_more_comments"]) && page.NextMaxID != ""
	moreMin := toBool(res["has_more_headload_comments"]) && page.NextMinID != ""
	page.HasMore = moreMax || moreMin
	return page
}

// MediaCommentsFirstPage loads the first page of comments for a media item. mediaID may be "mediaPk_userPk" or numeric media PK only (one extra info/ call to resolve owner).
func (c *Client) MediaCommentsFirstPage(mediaID string) (MediaCommentsPage, error) {
	return c.MediaCommentsFetch(mediaID, "", "")
}

// MediaCommentsFetch loads one comments page. For the first page pass empty maxID and minID. For pagination pass either next_max_id or next_min_id from the previous MediaCommentsPage (not both).
func (c *Client) MediaCommentsFetch(mediaID string, maxID, minID string) (MediaCommentsPage, error) {
	if maxID != "" && minID != "" {
		return MediaCommentsPage{}, fmt.Errorf("ig: pass at most one of maxID or minID")
	}
	rest, err := c.resolveMediaRESTID(mediaID)
	if err != nil {
		return MediaCommentsPage{}, err
	}
	params := url.Values{}
	params.Set("can_support_threading", "true")
	if maxID != "" {
		params.Set("max_id", maxID)
	}
	if minID != "" {
		params.Set("min_id", minID)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("media/%s/comments/", rest),
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return MediaCommentsPage{}, err
	}
	return parseMediaCommentsResponse(res), nil
}

// MediaCommentsNext fetches the next page using cursors from a previous page. Returns an error if HasMore is false or cursors are missing.
func (c *Client) MediaCommentsNext(mediaID string, prev MediaCommentsPage) (MediaCommentsPage, error) {
	if !prev.HasMore {
		return MediaCommentsPage{}, fmt.Errorf("ig: no more comments")
	}
	if prev.NextMaxID != "" {
		return c.MediaCommentsFetch(mediaID, prev.NextMaxID, "")
	}
	if prev.NextMinID != "" {
		return c.MediaCommentsFetch(mediaID, "", prev.NextMinID)
	}
	return MediaCommentsPage{}, fmt.Errorf("ig: missing comment cursor")
}
