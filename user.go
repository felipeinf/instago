package ig

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/felipeinf/instago/igerrors"
)

// UserIDFromUsername returns the numeric user id for username via UserInfoByUsername with caching enabled.
func (c *Client) UserIDFromUsername(username string) (int64, error) {
	u, err := c.UserInfoByUsername(username, true)
	if err != nil {
		return 0, err
	}
	return u.PK, nil
}

// UserInfoByUsernameGQL loads profile via public web_profile_info (GraphQL-backed web endpoint).
func (c *Client) UserInfoByUsernameGQL(username string) (User, error) {
	raw, err := c.WebProfileInfo(username)
	if err != nil {
		return User{}, err
	}
	data, ok := raw["data"].(map[string]any)
	if !ok {
		return User{}, fmt.Errorf("ig: web_profile_info missing data")
	}
	userNode, ok := data["user"].(map[string]any)
	if !ok {
		return User{}, fmt.Errorf("ig: web_profile_info missing user")
	}
	return extractUserGQLWebProfile(userNode)
}

// UserInfoByUsernameV1 calls users/{username}/usernameinfo/ and returns the raw response map alongside User.
func (c *Client) UserInfoByUsernameV1(username string) (User, map[string]any, error) {
	un := strings.ToLower(username)
	for _, r := range un {
		if r == '/' || r == '\\' || unicode.IsSpace(r) {
			return User{}, nil, fmt.Errorf("ig: invalid username")
		}
	}
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("users/%s/usernameinfo/", un),
		WithSignature: false,
	})
	if err != nil {
		var nf *igerrors.UserNotFound
		if igerrors.IsNotFound(err) || igerrors.AsUserNotFound(err, &nf) {
			return User{}, nil, &igerrors.UserNotFound{Username: un}
		}
		return User{}, nil, err
	}
	um, ok := res["user"].(map[string]any)
	if !ok {
		return User{}, res, fmt.Errorf("ig: usernameinfo missing user")
	}
	u, err := extractUserV1(um)
	return u, res, err
}

// UserInfoByUsername prefers web profile GraphQL, then falls back to usernameinfo. useCache skips refetch when a cached User exists; false clears cache for that name first.
func (c *Client) UserInfoByUsername(username string, useCache bool) (User, error) {
	un := strings.ToLower(username)
	if useCache {
		c.userMu.Lock()
		if pk, ok := c.usersByName[un]; ok {
			if u, ok2 := c.usersByPK[pk]; ok2 {
				c.userMu.Unlock()
				return u, nil
			}
		}
		c.userMu.Unlock()
	} else {
		c.userCacheDel(un)
	}
	u, err := c.UserInfoByUsernameGQL(un)
	if err != nil {
		var lr *igerrors.LoginRequired
		if errors.As(err, &lr) {
			if c.injectSessionIDToPublic() {
				u, err = c.UserInfoByUsernameGQL(un)
			}
		}
		if err != nil {
			u, _, err = c.UserInfoByUsernameV1(un)
			if err != nil {
				return User{}, err
			}
		}
	}
	c.userCacheSet(u)
	return u, nil
}

// UserInfo loads users/{id}/info/. useCache returns a previously stored User without a network call when true.
func (c *Client) UserInfo(userID int64, useCache bool) (User, error) {
	if useCache {
		if u, ok := c.userCacheGet(userID); ok {
			return u, nil
		}
	}
	params := url.Values{}
	params.Set("is_prefetch", "false")
	params.Set("entry_point", "self_profile")
	params.Set("from_module", "self_profile")
	params.Set("is_app_start", "false")
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("users/%d/info/", userID),
		Params:        params,
		WithSignature: true,
	})
	if err != nil {
		return User{}, err
	}
	um, ok := res["user"].(map[string]any)
	if !ok {
		return User{}, fmt.Errorf("ig: user info missing user")
	}
	u, err := extractUserV1(um)
	if err == nil {
		c.userCacheSet(u)
	}
	return u, err
}

// SearchUsersV1 calls users/search/ and returns up to count results plus the raw JSON map.
func (c *Client) SearchUsersV1(query string, count int) ([]UserShort, map[string]any, error) {
	params := url.Values{}
	params.Set("search_surface", "user_search_page")
	params.Set("timezone_offset", fmt.Sprintf("%d", c.timezoneOffset))
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("q", query)
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "users/search/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, nil, err
	}
	arr, ok := res["users"].([]any)
	if !ok {
		return nil, res, nil
	}
	out := make([]UserShort, 0, len(arr))
	for _, x := range arr {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		us, err := extractUserShort(m)
		if err != nil {
			continue
		}
		out = append(out, us)
	}
	return out, res, nil
}

func (c *Client) userCacheSet(u User) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	c.usersByPK[u.PK] = u
	c.usersByName[strings.ToLower(u.Username)] = u.PK
}

func (c *Client) userCacheGet(pk int64) (User, bool) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	u, ok := c.usersByPK[pk]
	return u, ok
}

func (c *Client) userCacheDel(username string) {
	c.userMu.Lock()
	defer c.userMu.Unlock()
	pk, ok := c.usersByName[strings.ToLower(username)]
	if !ok {
		return
	}
	delete(c.usersByPK, pk)
	delete(c.usersByName, strings.ToLower(username))
}
