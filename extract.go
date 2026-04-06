package ig

import (
	"encoding/json"
	"fmt"
	"sort"

	igenc "github.com/felipeinf/instago/encoding"
)

var mediaTypesGQL = map[string]int{
	"GraphImage":   1,
	"GraphVideo":   2,
	"GraphSidecar": 8,
	"StoryVideo":   2,
}

func jsonValue(root any, path ...any) any {
	cur := root
	for _, p := range path {
		switch k := p.(type) {
		case string:
			m, ok := cur.(map[string]any)
			if !ok {
				return nil
			}
			cur = m[k]
		case int:
			sl, ok := cur.([]any)
			if !ok || k < 0 || k >= len(sl) {
				return nil
			}
			cur = sl[k]
		default:
			return nil
		}
	}
	return cur
}

func bestGQLImageURL(candidates []any) string {
	type dim struct {
		area int
		url  string
	}
	var best dim
	for _, x := range candidates {
		m, ok := x.(map[string]any)
		if !ok {
			continue
		}
		w := int(toInt64(m["config_width"]))
		h := int(toInt64(m["config_height"]))
		if w == 0 {
			w = int(toInt64(m["width"]))
			h = int(toInt64(m["height"]))
		}
		u := toString(m["src"])
		if u == "" {
			u = toString(m["url"])
		}
		if w*h >= best.area {
			best = dim{w * h, u}
		}
	}
	return best.url
}

func extractMediaGql(media map[string]any) Media {
	tn, _ := media["__typename"].(string)
	mt := mediaTypesGQL[tn]
	pt := toString(media["product_type"])
	if mt == 2 && pt == "" {
		pt = "feed"
	}
	var candidates []any
	if dr, ok := media["display_resources"].([]any); ok {
		candidates = dr
	} else if tr, ok := media["thumbnail_resources"].([]any); ok {
		candidates = tr
	}
	thumb := bestGQLImageURL(candidates)
	if thumb == "" {
		thumb = toString(media["thumbnail_src"])
	}
	if mt == 8 {
		thumb = ""
	}
	videoURL := toString(media["video_url"])
	mediaID := toInt64(media["id"])
	var u UserShort
	if owner, ok := media["owner"].(map[string]any); ok {
		u, _ = extractUserShort(owner)
	}
	capText := ""
	if v := jsonValue(media, "edge_media_to_caption", "edges", 0, "node", "text"); v != nil {
		capText = toString(v)
	}
	code := toString(media["shortcode"])
	likeCount := 0
	if v := jsonValue(media, "edge_media_preview_like", "count"); v != nil {
		likeCount = int(toInt64(v))
	}
	commentCount := 0
	if v := jsonValue(media, "edge_media_to_comment", "count"); v != nil {
		commentCount = int(toInt64(v))
	}
	var utags []Usertag
	if edgeTag, ok := media["edge_media_to_tagged_user"].(map[string]any); ok {
		if edges, ok := edgeTag["edges"].([]any); ok {
			for _, e := range edges {
				em, _ := e.(map[string]any)
				node, _ := em["node"].(map[string]any)
				if node == nil {
					continue
				}
				utags = append(utags, extractUsertag(node))
			}
		}
	}
	out := Media{
		PK:           mediaID,
		ID:           fmt.Sprintf("%d_%d", mediaID, u.PK),
		Code:         code,
		MediaType:    mt,
		ProductType:  pt,
		TakenAt:      toInt64(media["taken_at_timestamp"]),
		ThumbnailURL: thumb,
		VideoURL:     videoURL,
		CaptionText:  capText,
		User:         u,
		LikeCount:    likeCount,
		CommentCount: commentCount,
		Usertags:     utags,
	}
	if mt == 8 {
		if esc, ok := media["edge_sidecar_to_children"].(map[string]any); ok {
			if edges, ok := esc["edges"].([]any); ok {
				for _, e := range edges {
					em, _ := e.(map[string]any)
					node, _ := em["node"].(map[string]any)
					if node != nil {
						out.CarouselMedia = append(out.CarouselMedia, extractMediaGql(node))
					}
				}
			}
		}
	}
	return out
}

func extractStoryGql(st map[string]any) Story {
	videoURL := ""
	if vr, ok := st["video_resources"].([]any); ok && len(vr) > 0 {
		type dim struct {
			area int
			url  string
		}
		var best dim
		for _, x := range vr {
			m, ok := x.(map[string]any)
			if !ok {
				continue
			}
			h := int(toInt64(m["config_height"]))
			w := int(toInt64(m["config_width"]))
			u := toString(m["src"])
			if w*h >= best.area {
				best = dim{w * h, u}
			}
		}
		videoURL = best.url
	}
	thumb := toString(st["display_url"])
	var u UserShort
	if owner, ok := st["owner"].(map[string]any); ok {
		u, _ = extractUserShort(owner)
	}
	idStr := toString(st["id"])
	pkNum := toInt64(st["id"])
	code := igenc.InstagramIDEncode(pkNum)
	isVideo := toBool(st["is_video"])
	mt := 1
	if isVideo {
		mt = 2
	}
	return Story{
		PK:           idStr,
		ID:           fmt.Sprintf("%s_%d", idStr, u.PK),
		Code:         code,
		MediaType:    mt,
		TakenAt:      toInt64(st["taken_at_timestamp"]),
		ThumbnailURL: thumb,
		VideoURL:     videoURL,
		ProductType:  "story",
		User:         u,
		Mentions:     nil,
		Hashtags:     nil,
		Locations:    nil,
	}
}

func extractUserShort(m map[string]any) (UserShort, error) {
	pk := toInt64(m["pk"])
	if pk == 0 {
		pk = toInt64(m["id"])
	}
	if pk == 0 {
		return UserShort{}, fmt.Errorf("extract: user without pk")
	}
	return UserShort{
		PK:            pk,
		Username:      toString(m["username"]),
		FullName:      toString(m["full_name"]),
		ProfilePicURL: toString(m["profile_pic_url"]),
		IsPrivate:     toBool(m["is_private"]),
		IsVerified:    toBool(m["is_verified"]),
	}, nil
}

func extractUserV1(m map[string]any) (User, error) {
	short, err := extractUserShort(m)
	if err != nil {
		return User{}, err
	}
	hd := ""
	if versions, ok := m["hd_profile_pic_versions"].([]any); ok && len(versions) > 0 {
		if last, ok := versions[len(versions)-1].(map[string]any); ok {
			hd = toString(last["url"])
		}
	}
	if hd == "" {
		if info, ok := m["hd_profile_pic_url_info"].(map[string]any); ok {
			hd = toString(info["url"])
		}
	}
	ext := m["external_url"]
	extStr := ""
	if ext != nil {
		extStr = toString(ext)
	}
	return User{
		UserShort:       short,
		Biography:       toString(m["biography"]),
		ExternalURL:     extStr,
		IsBusiness:      toBool(m["is_business"]),
		FollowerCount:   int(toInt64(m["follower_count"])),
		FollowingCount:  int(toInt64(m["following_count"])),
		MediaCount:      int(toInt64(m["media_count"])),
		ProfilePicURLHD: hd,
		PublicEmail:     toString(m["public_email"]),
		ContactPhone:    toString(m["contact_phone_number"]),
	}, nil
}

func extractUserGQLWebProfile(data map[string]any) (User, error) {
	edgeFollowed, _ := data["edge_followed_by"].(map[string]any)
	edgeFollow, _ := data["edge_follow"].(map[string]any)
	edgeMedia, _ := data["edge_owner_to_timeline_media"].(map[string]any)
	short := UserShort{
		PK:            toInt64(data["id"]),
		Username:      toString(data["username"]),
		FullName:      toString(data["full_name"]),
		ProfilePicURL: toString(data["profile_pic_url"]),
		IsPrivate:     toBool(data["is_private"]),
		IsVerified:    toBool(data["is_verified"]),
	}
	return User{
		UserShort:      short,
		Biography:      toString(data["biography"]),
		ExternalURL:    toString(data["external_url"]),
		IsBusiness:     toBool(data["is_business_account"]),
		FollowerCount:  int(toInt64(edgeFollowed["count"])),
		FollowingCount: int(toInt64(edgeFollow["count"])),
		MediaCount:     int(toInt64(edgeMedia["count"])),
	}, nil
}

func extractAccount(m map[string]any) Account {
	return Account{
		PK:          toString(m["pk"]),
		Username:    toString(m["username"]),
		FullName:    toString(m["full_name"]),
		Biography:   toString(m["biography"]),
		ExternalURL: toString(m["external_url"]),
		Email:       toString(m["email"]),
		PhoneNumber: toString(m["phone_number"]),
	}
}

func extractMediaV1(m map[string]any) Media {
	media := m
	videoURL := ""
	if vv, ok := media["video_versions"].([]any); ok && len(vv) > 0 {
		type dim struct {
			h, w int
			u    string
		}
		var best dim
		for _, x := range vv {
			vm, _ := x.(map[string]any)
			h := int(toInt64(vm["height"]))
			w := int(toInt64(vm["width"]))
			u := toString(vm["url"])
			if h*w >= best.h*best.w {
				best = dim{h, w, u}
			}
		}
		videoURL = best.u
	}
	thumb := ""
	if im, ok := media["image_versions2"].(map[string]any); ok {
		if cands, ok := im["candidates"].([]any); ok && len(cands) > 0 {
			type dim struct {
				h, w int
				u    string
			}
			var best dim
			for _, x := range cands {
				vm, _ := x.(map[string]any)
				h := int(toInt64(vm["height"]))
				w := int(toInt64(vm["width"]))
				u := toString(vm["url"])
				if h*w >= best.h*best.w {
					best = dim{h, w, u}
				}
			}
			thumb = best.u
		}
	}
	var u UserShort
	if um, ok := media["user"].(map[string]any); ok {
		u, _ = extractUserShort(um)
	}
	capText := ""
	if cap, ok := media["caption"].(map[string]any); ok {
		capText = toString(cap["text"])
	}
	mt := int(toInt64(media["media_type"]))
	pt := toString(media["product_type"])
	if mt == 2 && pt == "" {
		pt = "feed"
	}
	likeCount := int(toInt64(media["like_count"]))
	if likeCount == 0 {
		likeCount = int(toInt64(media["likes"]))
	}
	commentCount := int(toInt64(media["comment_count"]))
	var carousel []Media
	if carr, ok := media["carousel_media"].([]any); ok {
		for _, x := range carr {
			cm, _ := x.(map[string]any)
			if cm != nil {
				carousel = append(carousel, extractMediaV1(cm))
			}
		}
	}
	var utags []Usertag
	if ut, ok := media["usertags"].(map[string]any); ok {
		if in, ok := ut["in"].([]any); ok {
			for _, x := range in {
				m, _ := x.(map[string]any)
				if m != nil {
					utags = append(utags, extractUsertag(m))
				}
			}
		}
	}
	return Media{
		PK:            toInt64(media["pk"]),
		ID:            toString(media["id"]),
		Code:          toString(media["code"]),
		MediaType:     mt,
		ProductType:   pt,
		TakenAt:       toInt64(media["taken_at"]),
		ThumbnailURL:  thumb,
		VideoURL:      videoURL,
		CaptionText:   capText,
		User:          u,
		LikeCount:     likeCount,
		CommentCount:  commentCount,
		PlayCount:     int(toInt64(media["play_count"])),
		Usertags:      utags,
		CarouselMedia: carousel,
	}
}

func extractStoryV1(m map[string]any) Story {
	videoURL := ""
	if vv, ok := m["video_versions"].([]any); ok && len(vv) > 0 {
		sort.Slice(vv, func(i, j int) bool {
			a, _ := vv[i].(map[string]any)
			b, _ := vv[j].(map[string]any)
			return toInt64(a["height"])*toInt64(a["width"]) < toInt64(b["height"])*toInt64(b["width"])
		})
		last, _ := vv[len(vv)-1].(map[string]any)
		videoURL = toString(last["url"])
	}
	thumb := ""
	if im, ok := m["image_versions2"].(map[string]any); ok {
		if cands, ok := im["candidates"].([]any); ok && len(cands) > 0 {
			sort.Slice(cands, func(i, j int) bool {
				a, _ := cands[i].(map[string]any)
				b, _ := cands[j].(map[string]any)
				return toInt64(a["height"])*toInt64(a["width"]) < toInt64(b["height"])*toInt64(b["width"])
			})
			last, _ := cands[len(cands)-1].(map[string]any)
			thumb = toString(last["url"])
		}
	}
	mt := int(toInt64(m["media_type"]))
	pt := toString(m["product_type"])
	if mt == 2 && pt == "" {
		pt = "story"
	}
	var u UserShort
	if um, ok := m["user"].(map[string]any); ok {
		u, _ = extractUserShort(um)
	}
	var mentions []StoryMention
	if rm, ok := m["reel_mentions"].([]any); ok {
		for _, x := range rm {
			mm, _ := x.(map[string]any)
			if mm == nil {
				continue
			}
			var mus UserShort
			if um, ok := mm["user"].(map[string]any); ok {
				mus, _ = extractUserShort(um)
			}
			mentions = append(mentions, StoryMention{
				User: mus, X: toFloat(mm["x"]), Y: toFloat(mm["y"]),
				Width: toFloat(mm["width"]), Height: toFloat(mm["height"]), Rotate: toFloat(mm["rotation"]),
			})
		}
	}
	var hashtags []StoryHashtag
	if sh, ok := m["story_hashtags"].([]any); ok {
		for _, x := range sh {
			mm, _ := x.(map[string]any)
			if mm == nil {
				continue
			}
			hm, _ := mm["hashtag"].(map[string]any)
			var h Hashtag
			if hm != nil {
				h = Hashtag{ID: toInt64(hm["id"]), Name: toString(hm["name"])}
			}
			hashtags = append(hashtags, StoryHashtag{
				Hashtag: h, X: toFloat(mm["x"]), Y: toFloat(mm["y"]),
				Width: toFloat(mm["width"]), Height: toFloat(mm["height"]), Rotate: toFloat(mm["rotation"]),
			})
		}
	}
	var locs []StoryLocation
	if sl, ok := m["story_locations"].([]any); ok {
		for _, x := range sl {
			mm, _ := x.(map[string]any)
			if mm == nil {
				continue
			}
			lm, _ := mm["location"].(map[string]any)
			loc := StoryLocation{
				X: toFloat(mm["x"]), Y: toFloat(mm["y"]),
				Width: toFloat(mm["width"]), Height: toFloat(mm["height"]), Rotate: toFloat(mm["rotation"]),
			}
			if lm != nil {
				loc.LocationPK = toInt64(lm["pk"])
				if loc.LocationPK == 0 {
					loc.LocationPK = toInt64(lm["location_id"])
				}
				loc.Name = toString(lm["name"])
				loc.Address = toString(lm["address"])
				loc.City = toString(lm["city"])
				loc.ExternalID = toString(lm["external_id"])
				loc.ExternalSrc = toString(lm["external_id_source"])
			}
			locs = append(locs, loc)
		}
	}
	return Story{
		PK:           toString(m["pk"]),
		ID:           toString(m["id"]),
		Code:         toString(m["code"]),
		MediaType:    mt,
		TakenAt:      toInt64(m["taken_at"]),
		ThumbnailURL: thumb,
		VideoURL:     videoURL,
		ProductType:  pt,
		User:         u,
		Mentions:     mentions,
		Hashtags:     hashtags,
		Locations:    locs,
	}
}

func extractUsertag(m map[string]any) Usertag {
	x, y := toFloat(m["x"]), toFloat(m["y"])
	if pos, ok := m["position"].([]any); ok && len(pos) >= 2 {
		x = toFloat(pos[0])
		y = toFloat(pos[1])
	}
	var u UserShort
	if um, ok := m["user"].(map[string]any); ok {
		u, _ = extractUserShort(um)
	}
	return Usertag{User: u, X: x, Y: y}
}

func toSliceMap(key string, m map[string]any) []map[string]any {
	v, ok := m[key].([]any)
	if !ok {
		return nil
	}
	var out []map[string]any
	for _, x := range v {
		mm, ok := x.(map[string]any)
		if ok {
			out = append(out, mm)
		}
	}
	return out
}

func bestDirectImageURL(m map[string]any) string {
	im, ok := m["image_versions2"].(map[string]any)
	if !ok {
		return ""
	}
	cands, ok := im["candidates"].([]any)
	if !ok || len(cands) == 0 {
		return ""
	}
	type dim struct{ area int; u string }
	var best dim
	for _, x := range cands {
		vm, _ := x.(map[string]any)
		if vm == nil {
			continue
		}
		h := int(toInt64(vm["height"]))
		w := int(toInt64(vm["width"]))
		u := toString(vm["url"])
		if w*h >= best.area {
			best = dim{w * h, u}
		}
	}
	return best.u
}

func bestDirectVideoURL(m map[string]any) string {
	vv, ok := m["video_versions"].([]any)
	if !ok || len(vv) == 0 {
		return ""
	}
	type dim struct{ area int; u string }
	var best dim
	for _, x := range vv {
		vm, _ := x.(map[string]any)
		if vm == nil {
			continue
		}
		h := int(toInt64(vm["height"]))
		w := int(toInt64(vm["width"]))
		u := toString(vm["url"])
		if w*h >= best.area {
			best = dim{w * h, u}
		}
	}
	return best.u
}

func extractDirectMessage(m map[string]any, threadID string) DirectMessage {
	id := toString(m["item_id"])
	if id == "" {
		id = toString(m["id"])
	}
	ts := toInt64(m["timestamp"])
	text := toString(m["text"])
	itemType := toString(m["item_type"])
	dm := DirectMessage{
		ID:          id,
		ThreadID:    threadID,
		UserID:      toString(m["user_id"]),
		TimestampUS: ts,
		Text:        text,
		ItemType:    itemType,
	}
	if med, ok := m["media"].(map[string]any); ok {
		dm.MediaURL = bestDirectImageURL(med)
		if dm.MediaURL == "" {
			dm.MediaURL = bestDirectVideoURL(med)
		}
		dm.MediaVideoURL = bestDirectVideoURL(med)
	}
	if vm, ok := m["visual_media"].(map[string]any); ok {
		if inner, ok := vm["media"].(map[string]any); ok {
			dm.VisualPhotoURL = bestDirectImageURL(inner)
			dm.VisualVideoURL = bestDirectVideoURL(inner)
		}
	}
	if vm, ok := m["voice_media"].(map[string]any); ok {
		if med, ok := vm["media"].(map[string]any); ok {
			dm.MediaURL = bestDirectImageURL(med)
			dm.MediaVideoURL = bestDirectVideoURL(med)
		}
	}
	return dm
}

func extractDirectThreadMap(m map[string]any) DirectThread {
	pk := toString(m["thread_v2_id"])
	id := toString(m["thread_id"])
	var users []UserShort
	for _, u := range toSliceMap("users", m) {
		if us, err := extractUserShort(u); err == nil {
			users = append(users, us)
		}
	}
	var msgs []DirectMessage
	for _, it := range toSliceMap("items", m) {
		msgs = append(msgs, extractDirectMessage(it, id))
	}
	last := toInt64(m["last_activity_at"])
	if last > 1_000_000_000_000 {
		last /= 1_000_000
	}
	title := toString(m["thread_title"])
	if title == "" {
		title = toString(m["thread_name"])
	}
	return DirectThread{
		PK: pk, ID: id, Users: users, Messages: msgs,
		Title: title, ThreadType: toString(m["thread_type"]),
		LastActivity: last,
	}
}

func extractComment(m map[string]any) Comment {
	var u UserShort
	if um, ok := m["user"].(map[string]any); ok {
		u, _ = extractUserShort(um)
	}
	pk := toInt64(m["pk"])
	if pk == 0 {
		pk = toInt64(m["id"])
	}
	created := toInt64(m["created_at"])
	if created == 0 {
		created = toInt64(m["created_at_utc"])
	}
	lc := int(toInt64(m["comment_like_count"]))
	if lc == 0 {
		lc = int(toInt64(m["like_count"]))
	}
	return Comment{
		PK: pk, Text: toString(m["text"]), User: u,
		CreatedAt: created, LikeCount: lc,
	}
}

func extractHashtagV1(m map[string]any) Hashtag {
	return Hashtag{
		ID:            toInt64(m["id"]),
		Name:          toString(m["name"]),
		MediaCount:    int(toInt64(m["media_count"])),
		ProfilePicURL: toString(m["profile_pic_url"]),
	}
}

func extractTrack(m map[string]any) Track {
	return Track{
		ID:    toString(m["id"]),
		Title: toString(m["title"]),
		URI:   toString(m["uri"]),
	}
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	case float64:
		return fmt.Sprintf("%.0f", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	default:
		return fmt.Sprint(x)
	}
}

func toInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case json.Number:
		i, _ := x.Int64()
		return i
	case string:
		var i int64
		_, _ = fmt.Sscanf(x, "%d", &i)
		return i
	default:
		return 0
	}
}

func toBool(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		return x == "1" || x == "true"
	default:
		return false
	}
}

func toFloat(v any) float64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	default:
		return 0
	}
}
