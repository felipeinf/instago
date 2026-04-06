package ig

import (
	"fmt"
	"net/url"
	"strconv"
)

// FbsearchTopsearchFlat calls fbsearch/topsearch_flat/ and returns the list slice from the response.
func (c *Client) FbsearchTopsearchFlat(query string) ([]any, error) {
	params := url.Values{}
	params.Set("search_surface", "top_search_page")
	params.Set("context", "blended")
	params.Set("timezone_offset", strconv.Itoa(c.timezoneOffset))
	params.Set("count", "30")
	params.Set("query", query)
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "fbsearch/topsearch_flat/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	list, _ := res["list"].([]any)
	return list, nil
}

// SearchUsersFB searches users via users/search/ (FB-style surface).
func (c *Client) SearchUsersFB(query string) ([]UserShort, error) {
	params := url.Values{}
	params.Set("search_surface", "user_search_page")
	params.Set("timezone_offset", strconv.Itoa(c.timezoneOffset))
	params.Set("count", "30")
	params.Set("q", query)
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "users/search/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	arr, _ := res["users"].([]any)
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
	return out, nil
}

// SearchHashtags calls tags/search/ and returns hashtag results.
func (c *Client) SearchHashtags(query string) ([]Hashtag, error) {
	params := url.Values{}
	params.Set("search_surface", "hashtag_search_page")
	params.Set("timezone_offset", strconv.Itoa(c.timezoneOffset))
	params.Set("count", "30")
	params.Set("q", query)
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "tags/search/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	arr, _ := res["results"].([]any)
	out := make([]Hashtag, 0, len(arr))
	for _, x := range arr {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, extractHashtagV1(m))
	}
	return out, nil
}

// SearchMusic calls music/audio_global_search/.
func (c *Client) SearchMusic(query string) ([]Track, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("browse_session_id", c.generateUUID("", ""))
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "music/audio_global_search/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	arr, _ := res["items"].([]any)
	out := make([]Track, 0, len(arr))
	for _, x := range arr {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		tr, ok := m["track"].(map[string]any)
		if !ok {
			continue
		}
		out = append(out, extractTrack(tr))
	}
	return out, nil
}

// FbsearchPlaces searches places near lat/lng via fbsearch/places/.
func (c *Client) FbsearchPlaces(query string, lat, lng float64) ([]map[string]any, error) {
	params := url.Values{}
	params.Set("search_surface", "places_search_page")
	params.Set("timezone_offset", strconv.Itoa(c.timezoneOffset))
	params.Set("lat", fmt.Sprintf("%g", lat))
	params.Set("lng", fmt.Sprintf("%g", lng))
	params.Set("count", "30")
	params.Set("query", query)
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "fbsearch/places/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	items, _ := res["items"].([]any)
	out := make([]map[string]any, 0, len(items))
	for _, x := range items {
		m, ok := x.(map[string]any)
		if ok {
			out = append(out, m)
		}
	}
	return out, nil
}

// FbsearchRecent returns recent searches from fbsearch/recent_searches/.
func (c *Client) FbsearchRecent() ([]any, error) {
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "fbsearch/recent_searches/",
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	if st, _ := res["status"].(string); st != "" && st != "ok" {
		return nil, fmt.Errorf("ig: recent_searches status %q", st)
	}
	recent, _ := res["recent"].([]any)
	return recent, nil
}

// FbsearchSuggestedProfiles returns account recommendations for target_user_id via fbsearch/accounts_recs/.
func (c *Client) FbsearchSuggestedProfiles(userID int64) ([]UserShort, error) {
	params := url.Values{}
	params.Set("target_user_id", strconv.FormatInt(userID, 10))
	params.Set("include_friendship_status", "true")
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      "fbsearch/accounts_recs/",
		Params:        params,
		WithSignature: false,
	})
	if err != nil {
		return nil, err
	}
	arr, _ := res["users"].([]any)
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
	return out, nil
}
