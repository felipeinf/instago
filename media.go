package ig

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/felipeinf/instago/igerrors"
)

const userTimelineGQLHash = "e7e2f4da4b02303f74f0841279e52d76"

// RankToken returns a token used when paginating user feeds (user id and client UUID).
func (c *Client) RankToken() string {
	return fmt.Sprintf("%d_%s", c.userID(), c.uuids.UUID)
}

// UserMediasPaginatedV1 fetches one page from feed/user/{id}/; maxID is the next cursor from the previous call.
func (c *Client) UserMediasPaginatedV1(userID int64, amount int, maxID string) ([]Media, string, error) {
	params := url.Values{}
	if maxID != "" {
		params.Set("max_id", maxID)
	}
	params.Set("count", strconv.Itoa(amount))
	params.Set("rank_token", c.RankToken())
	params.Set("ranked_content", "true")
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("feed/user/%d/", userID),
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, "", err
	}
	items, _ := res["items"].([]any)
	out := make([]Media, 0, len(items))
	for _, x := range items {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, extractMediaV1(m))
	}
	next, _ := res["next_max_id"].(string)
	if amount > 0 && len(out) > amount {
		out = out[:amount]
	}
	return out, next, nil
}

// UserMediasV1 collects up to amount posts using the private REST feed, following next_max_id until done.
func (c *Client) UserMediasV1(userID int64, amount int) ([]Media, error) {
	var all []Media
	next := ""
	page := 33
	if amount > 0 && amount < page {
		page = amount
	}
	for {
		chunk, nxt, err := c.UserMediasPaginatedV1(userID, page, next)
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

// UserMediasPaginatedGQL fetches one GraphQL timeline page; endCursor comes from the previous page's page_info.
func (c *Client) UserMediasPaginatedGQL(userID int64, pageAmount int, endCursor string) ([]Media, string, error) {
	first := 50
	if pageAmount > 0 && pageAmount < 50 {
		first = pageAmount
	}
	vars := map[string]any{
		"id":    userID,
		"first": first,
	}
	if endCursor != "" {
		vars["after"] = endCursor
	}
	data, err := c.PublicGraphqlRequest(vars, userTimelineGQLHash)
	if err != nil {
		return nil, "", err
	}
	userNode, _ := data["user"].(map[string]any)
	if userNode == nil {
		return nil, "", fmt.Errorf("ig: gql user unavailable")
	}
	edge, _ := userNode["edge_owner_to_timeline_media"].(map[string]any)
	if edge == nil {
		return nil, "", nil
	}
	pageInfo, _ := edge["page_info"].(map[string]any)
	edges, _ := edge["edges"].([]any)
	next := ""
	if pageInfo != nil {
		next, _ = pageInfo["end_cursor"].(string)
	}
	out := make([]Media, 0, len(edges))
	for _, x := range edges {
		e, ok := x.(map[string]any)
		if !ok {
			continue
		}
		node, _ := e["node"].(map[string]any)
		if node == nil {
			continue
		}
		out = append(out, extractMediaGql(node))
	}
	if pageAmount > 0 && len(out) > pageAmount {
		out = out[:pageAmount]
	}
	return out, next, nil
}

// UserMediasGQL paginates the public GraphQL user timeline until amount items or no next cursor. sleepSec is the delay between pages; 0 picks a short random sleep.
func (c *Client) UserMediasGQL(userID int64, amount int, sleepSec int) ([]Media, error) {
	var all []Media
	cursor := ""
	for {
		remain := 0
		if amount > 0 {
			remain = amount - len(all)
			if remain <= 0 {
				break
			}
		}
		page, next, err := c.UserMediasPaginatedGQL(userID, remain, cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, page...)
		if next == "" || len(page) == 0 {
			break
		}
		if amount > 0 && len(all) >= amount {
			break
		}
		cursor = next
		sleep := sleepSec
		if sleep <= 0 {
			sleep = 1 + c.rng.Intn(3)
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	if amount > 0 && len(all) > amount {
		all = all[:amount]
	}
	return all, nil
}

func (c *Client) userMediasTryGQL(userID int64, amount int, sleepSec int) ([]Media, error) {
	out, err := c.UserMediasGQL(userID, amount, sleepSec)
	if err != nil {
		var lr *igerrors.LoginRequired
		if errors.As(err, &lr) && c.injectSessionIDToPublic() {
			return c.UserMediasGQL(userID, amount, sleepSec)
		}
		return nil, err
	}
	return out, nil
}

// UserMediasWithSleep tries GraphQL pagination first (with sleepSec between pages), then UserMediasV1 if GraphQL fails for non-login reasons.
func (c *Client) UserMediasWithSleep(userID int64, amount int, sleepSec int) ([]Media, error) {
	gql, err := c.userMediasTryGQL(userID, amount, sleepSec)
	if err == nil {
		return gql, nil
	}
	var lr *igerrors.LoginRequired
	if errors.As(err, &lr) {
		return nil, err
	}
	return c.UserMediasV1(userID, amount)
}

// UserMedias is the high-level helper: GraphQL with automatic backoff, then REST fallback.
func (c *Client) UserMedias(userID int64, amount int) ([]Media, error) {
	return c.UserMediasWithSleep(userID, amount, 0)
}
