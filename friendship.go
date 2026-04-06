package ig

import (
	"fmt"
	"net/url"
	"strconv"
)

func friendshipFromShowResponse(res map[string]any, userID int64) Friendship {
	return Friendship{
		UserID:           userID,
		Following:        toBool(res["following"]),
		FollowedBy:       toBool(res["followed_by"]),
		IncomingRequest:  toBool(res["incoming_request"]),
		OutgoingRequest:  toBool(res["outgoing_request"]),
		IsPrivate:        toBool(res["is_private"]),
		IsRestricted:     toBool(res["is_restricted"]),
		Blocking:         toBool(res["blocking"]),
	}
}

func friendshipFromStatus(res map[string]any, userID int64) (Friendship, error) {
	fs, _ := res["friendship_status"].(map[string]any)
	if fs == nil {
		return Friendship{}, fmt.Errorf("ig: friendship_status missing")
	}
	return Friendship{
		UserID:           userID,
		Following:        toBool(fs["following"]),
		FollowedBy:       toBool(fs["followed_by"]),
		IncomingRequest:  toBool(fs["incoming_request"]),
		OutgoingRequest:  toBool(fs["outgoing_request"]),
		IsPrivate:        toBool(fs["is_private"]),
		IsRestricted:     toBool(fs["is_restricted"]),
		Blocking:         toBool(fs["blocking"]),
	}, nil
}

// FriendshipWith returns relationship flags for the logged-in user vs targetUserID (friendships/show/).
func (c *Client) FriendshipWith(targetUserID int64) (Friendship, error) {
	if c.userID() == 0 {
		return Friendship{}, fmt.Errorf("ig: login required")
	}
	params := url.Values{}
	params.Set("is_external_deeplink_profile_view", "false")
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("friendships/show/%d/", targetUserID),
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return Friendship{}, err
	}
	return friendshipFromShowResponse(res, targetUserID), nil
}

// Follow follows targetUserID.
func (c *Client) Follow(targetUserID int64) (Friendship, error) {
	if c.userID() == 0 {
		return Friendship{}, fmt.Errorf("ig: login required")
	}
	idStr := strconv.FormatInt(targetUserID, 10)
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("friendships/create/%s/", idStr),
		Data:          c.withActionData(map[string]any{"user_id": idStr}),
		WithSignature: true,
	})
	if err != nil {
		return Friendship{}, err
	}
	return friendshipFromStatus(res, targetUserID)
}

// Unfollow unfollows targetUserID.
func (c *Client) Unfollow(targetUserID int64) (Friendship, error) {
	if c.userID() == 0 {
		return Friendship{}, fmt.Errorf("ig: login required")
	}
	idStr := strconv.FormatInt(targetUserID, 10)
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("friendships/destroy/%s/", idStr),
		Data:          c.withActionData(map[string]any{"user_id": idStr}),
		WithSignature: true,
	})
	if err != nil {
		return Friendship{}, err
	}
	return friendshipFromStatus(res, targetUserID)
}

// MutualFriendsPage is one page of mutual followers (friendships/{id}/mutual_friends/). Pass empty maxID for the first page.
func (c *Client) MutualFriendsPage(targetUserID int64, maxID string) ([]UserShort, string, error) {
	if c.userID() == 0 {
		return nil, "", fmt.Errorf("ig: login required")
	}
	params := url.Values{}
	params.Set("rank_token", c.RankToken())
	if maxID != "" {
		params.Set("max_id", maxID)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("friendships/%d/mutual_friends/", targetUserID),
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, "", err
	}
	var out []UserShort
	if arr, ok := res["users"].([]any); ok {
		for _, x := range arr {
			if um, ok := x.(map[string]any); ok {
				u, err := extractUserShort(um)
				if err == nil {
					out = append(out, u)
				}
			}
		}
	}
	next := toString(res["next_max_id"])
	return out, next, nil
}

// UserFollowersPage returns one page of followers for userID (friendships/{userID}/followers/). Pass empty maxID for the first page.
func (c *Client) UserFollowersPage(userID int64, maxID string) ([]UserShort, string, error) {
	if c.userID() == 0 {
		return nil, "", fmt.Errorf("ig: login required")
	}
	params := url.Values{}
	params.Set("rank_token", c.RankToken())
	params.Set("search_surface", "followers_list")
	if maxID != "" {
		params.Set("max_id", maxID)
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("friendships/%d/followers/", userID),
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, "", err
	}
	var out []UserShort
	if arr, ok := res["users"].([]any); ok {
		for _, x := range arr {
			if um, ok := x.(map[string]any); ok {
				u, err := extractUserShort(um)
				if err == nil {
					out = append(out, u)
				}
			}
		}
	}
	next := toString(res["next_max_id"])
	return out, next, nil
}
