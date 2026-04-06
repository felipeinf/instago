package ig

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	igcfg "github.com/felipeinf/instago/config"
	"github.com/felipeinf/instago/igerrors"
)

const userReelsGQLHash = "303a4ae99711322310f25250d988f3b7"

// UserStoriesV1 loads stories from the private feed/user/{id}/story/ endpoint.
func (c *Client) UserStoriesV1(userID int64, amount int) ([]Story, error) {
	capJSON, _ := json.Marshal(igcfg.SupportedCapabilities)
	params := url.Values{}
	params.Set("supported_capabilities_new", string(capJSON))
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("feed/user/%d/story/", userID),
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	reel, _ := res["reel"].(map[string]any)
	if reel == nil {
		return nil, nil
	}
	items, _ := reel["items"].([]any)
	out := make([]Story, 0, len(items))
	for _, x := range items {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, extractStoryV1(m))
	}
	if amount > 0 && len(out) > amount {
		out = out[:amount]
	}
	return out, nil
}

func (c *Client) userStoriesGQLSingle(userID int64, amount int) ([]Story, error) {
	c.injectSessionIDToPublic()
	data, err := c.PublicGraphqlRequest(map[string]any{
		"reel_ids":             []any{userID},
		"precomposed_overlay":  false,
	}, userReelsGQLHash)
	if err != nil {
		return nil, err
	}
	reels, _ := data["reels_media"].([]any)
	var items []any
	for _, x := range reels {
		rm, ok := x.(map[string]any)
		if !ok {
			continue
		}
		owner, _ := rm["owner"].(map[string]any)
		if toInt64(owner["id"]) != userID && toInt64(owner["pk"]) != userID {
			continue
		}
		items, _ = rm["items"].([]any)
		break
	}
	out := make([]Story, 0, len(items))
	for _, it := range items {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, extractStoryGql(m))
	}
	if amount > 0 && len(out) > amount {
		out = out[:amount]
	}
	return out, nil
}

// UserStories prefers public GraphQL reels_media; on login_required it may inject sessionid and retry, then falls back to UserStoriesV1.
func (c *Client) UserStories(userID int64, amount int) ([]Story, error) {
	gql, err := c.userStoriesGQLSingle(userID, amount)
	if err == nil {
		return gql, nil
	}
	var lr *igerrors.LoginRequired
	if errors.As(err, &lr) && c.injectSessionIDToPublic() {
		gql, err = c.userStoriesGQLSingle(userID, amount)
		if err == nil {
			return gql, nil
		}
	}
	return c.UserStoriesV1(userID, amount)
}

// StoryInfo parses a story primary key of form "{mediaPK}_{userID}", loads that user's stories, and returns the matching item.
func (c *Client) StoryInfo(storyPK string) (Story, error) {
	idx := strings.LastIndex(storyPK, "_")
	if idx <= 0 || idx >= len(storyPK)-1 {
		return Story{}, fmt.Errorf("ig: invalid story pk")
	}
	targetPK := storyPK[:idx]
	uid, err := strconv.ParseInt(storyPK[idx+1:], 10, 64)
	if err != nil {
		return Story{}, err
	}
	stories, err := c.UserStories(uid, 0)
	if err != nil {
		return Story{}, err
	}
	for _, s := range stories {
		if s.PK == targetPK || strings.HasPrefix(s.ID, targetPK+"_") {
			return s, nil
		}
	}
	return Story{}, fmt.Errorf("ig: story not found")
}
