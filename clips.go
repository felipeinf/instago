package ig

import (
	"fmt"
	"time"
)

// UserClipsPaginatedV1 returns one page of reels for a user (POST clips/user/). pageSize is passed as page_size (app default is often 12; use 0 for 12).
func (c *Client) UserClipsPaginatedV1(userID int64, pageSize int, maxID string) ([]Media, string, error) {
	if c.userID() == 0 {
		return nil, "", fmt.Errorf("ig: login required")
	}
	ps := pageSize
	if ps <= 0 {
		ps = 12
	}
	data := map[string]any{
		"target_user_id":     userID,
		"max_id":             maxID,
		"page_size":          ps,
		"include_feed_video": "true",
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "clips/user/",
		Data:          data,
		WithSignature: true,
	})
	if err != nil {
		return nil, "", err
	}
	items, _ := res["items"].([]any)
	out := make([]Media, 0, len(items))
	for _, x := range items {
		row, ok := x.(map[string]any)
		if !ok {
			continue
		}
		med, _ := row["media"].(map[string]any)
		if med == nil {
			continue
		}
		out = append(out, extractMediaV1(med))
	}
	var next string
	if pi, ok := res["paging_info"].(map[string]any); ok {
		next = toString(pi["max_id"])
	}
	if pageSize > 0 && len(out) > pageSize {
		out = out[:pageSize]
	}
	return out, next, nil
}

// UserClipsV1 collects up to amount reels using max_id pagination. amount 0 means fetch until exhaustion.
func (c *Client) UserClipsV1(userID int64, amount int) ([]Media, error) {
	var all []Media
	next := ""
	page := 12
	if amount > 0 && amount < page {
		page = amount
	}
	for {
		chunk, nxt, err := c.UserClipsPaginatedV1(userID, page, next)
		if err != nil {
			return nil, err
		}
		all = append(all, chunk...)
		if nxt == "" {
			break
		}
		if amount > 0 && len(all) >= amount {
			break
		}
		next = nxt
		time.Sleep(c.requestTimeout)
	}
	if amount > 0 && len(all) > amount {
		all = all[:amount]
	}
	return all, nil
}
