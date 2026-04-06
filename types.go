package ig

// Structs are ordered: session snapshot, profiles, feed content, search. Session JSON I/O is in settings.go.

// UUIDs holds client-generated identifiers sent with private API requests.
type UUIDs struct {
	PhoneID         string `json:"phone_id"`
	UUID            string `json:"uuid"`
	ClientSessionID string `json:"client_session_id"`
	AdvertisingID   string `json:"advertising_id"`
	AndroidDeviceID string `json:"android_device_id"`
	RequestID       string `json:"request_id"`
	TraySessionID   string `json:"tray_session_id"`
}

// DeviceSettings describes the emulated Android device and app build (version, bloks id, screen, CPU).
type DeviceSettings struct {
	AndroidVersion    int    `json:"android_version"`
	AndroidRelease    string `json:"android_release"`
	DPI               string `json:"dpi"`
	Resolution        string `json:"resolution"`
	Manufacturer      string `json:"manufacturer"`
	Device            string `json:"device"`
	Model             string `json:"model"`
	CPU               string `json:"cpu"`
	AppVersion        string `json:"app_version,omitempty"`
	VersionCode       string `json:"version_code,omitempty"`
	BloksVersioningID string `json:"bloks_versioning_id,omitempty"`
}

// Settings is the on-disk session snapshot: UUIDs, cookies, authorization blob, device, and locale.
type Settings struct {
	UUIDs             UUIDs             `json:"uuids"`
	Mid               string            `json:"mid"`
	IgURur            *string           `json:"ig_u_rur"`
	IgWWWClaim        *string           `json:"ig_www_claim"`
	AuthorizationData map[string]string `json:"authorization_data"`
	Cookies           map[string]string `json:"cookies"`
	LastLogin         *float64          `json:"last_login"`
	DeviceSettings    DeviceSettings    `json:"device_settings"`
	UserAgent         string            `json:"user_agent"`
	Country           string            `json:"country"`
	CountryCode       int               `json:"country_code"`
	Locale            string            `json:"locale"`
	TimezoneOffset    int               `json:"timezone_offset"`
}

// UserShort is a compact user record (search results, media owner, story owner).
type UserShort struct {
	PK            int64  `json:"pk"`
	Username      string `json:"username"`
	FullName      string `json:"full_name"`
	ProfilePicURL string `json:"profile_pic_url"`
	IsPrivate     bool   `json:"is_private"`
	IsVerified    bool   `json:"is_verified"`
}

// User is a full public profile returned by user info endpoints.
type User struct {
	UserShort
	Biography       string `json:"biography"`
	ExternalURL     string `json:"external_url"`
	IsBusiness      bool   `json:"is_business"`
	FollowerCount   int    `json:"follower_count"`
	FollowingCount  int    `json:"following_count"`
	MediaCount      int    `json:"media_count"`
	ProfilePicURLHD string `json:"profile_pic_url_hd"`
	PublicEmail     string `json:"public_email"`
	ContactPhone    string `json:"contact_phone_number"`
}

// Account is the logged-in account from accounts/current_user.
type Account struct {
	PK          string `json:"pk"`
	Username    string `json:"username"`
	FullName    string `json:"full_name"`
	Biography   string `json:"biography"`
	ExternalURL string `json:"external_url"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

// Usertag is an @mention on a photo or video (position in frame).
type Usertag struct {
	User UserShort `json:"user"`
	X    float64   `json:"x"`
	Y    float64   `json:"y"`
}

// Media is a feed post (photo, video, or album) with caption and owner.
type Media struct {
	PK           int64     `json:"pk"`
	ID           string    `json:"id"`
	Code         string    `json:"code"`
	MediaType    int       `json:"media_type"`
	ProductType  string    `json:"product_type"`
	TakenAt      int64     `json:"taken_at"`
	ThumbnailURL string    `json:"thumbnail_url"`
	VideoURL     string    `json:"video_url"`
	CaptionText  string    `json:"caption_text"`
	User         UserShort `json:"user"`
	LikeCount    int       `json:"like_count"`
	CommentCount int       `json:"comment_count"`
	Usertags     []Usertag `json:"usertags"`
	PlayCount    int       `json:"play_count"`
	CarouselMedia []Media  `json:"carousel_media,omitempty"`
}

// Comment is a comment on a media item.
type Comment struct {
	PK        int64     `json:"pk"`
	Text      string    `json:"text"`
	User      UserShort `json:"user"`
	CreatedAt int64     `json:"created_at"`
	LikeCount int       `json:"like_count"`
}

// StoryMention is an @mention sticker on a story (private v1).
type StoryMention struct {
	User   UserShort `json:"user"`
	X      float64   `json:"x"`
	Y      float64   `json:"y"`
	Width  float64   `json:"width"`
	Height float64   `json:"height"`
	Rotate float64   `json:"rotation"`
}

// StoryHashtag is a hashtag sticker on a story (private v1).
type StoryHashtag struct {
	Hashtag Hashtag `json:"hashtag"`
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
	Rotate  float64 `json:"rotation"`
}

// StoryLocation is a location sticker on a story (private v1).
type StoryLocation struct {
	LocationPK   int64  `json:"location_pk"`
	Name         string `json:"name"`
	Address      string `json:"address"`
	City         string `json:"city"`
	ExternalID   string `json:"external_id"`
	ExternalSrc  string `json:"external_id_source"`
	X            float64
	Y            float64
	Width        float64
	Height       float64
	Rotate       float64
}

// Story is a single story item (image or video) with owner metadata.
type Story struct {
	PK           string          `json:"pk"`
	ID           string          `json:"id"`
	Code         string          `json:"code"`
	MediaType    int             `json:"media_type"`
	TakenAt      int64           `json:"taken_at"`
	ThumbnailURL string          `json:"thumbnail_url"`
	VideoURL     string          `json:"video_url"`
	ProductType  string          `json:"product_type"`
	User         UserShort       `json:"user"`
	Mentions     []StoryMention  `json:"mentions"`
	Hashtags     []StoryHashtag  `json:"hashtags"`
	Locations    []StoryLocation `json:"locations"`
}

// DirectThread is a DM thread (inbox or full fetch).
type DirectThread struct {
	PK           string           `json:"pk"`
	ID           string           `json:"id"`
	Users        []UserShort      `json:"users"`
	Messages     []DirectMessage  `json:"messages"`
	Title        string           `json:"thread_title"`
	ThreadType   string           `json:"thread_type"`
	Inviter      *UserShort       `json:"inviter,omitempty"`
	LastActivity int64            `json:"last_activity_at"`
}

// DirectMessage is one item in a DM thread.
type DirectMessage struct {
	ID             string `json:"id"`
	ThreadID       string `json:"thread_id"`
	UserID         string `json:"user_id"`
	TimestampUS    int64  `json:"timestamp_us"`
	Text           string `json:"text"`
	ItemType       string `json:"item_type"`
	MediaURL       string `json:"media_url"`
	MediaVideoURL  string `json:"media_video_url"`
	VisualPhotoURL string `json:"visual_photo_url"`
	VisualVideoURL string `json:"visual_video_url"`
}

// Friendship describes relationship between the logged-in user and another user.
type Friendship struct {
	UserID         int64 `json:"user_id"`
	Following      bool  `json:"following"`
	FollowedBy     bool  `json:"followed_by"`
	IncomingRequest bool `json:"incoming_request"`
	OutgoingRequest bool `json:"outgoing_request"`
	IsPrivate      bool  `json:"is_private"`
	IsRestricted   bool  `json:"is_restricted"`
	Blocking       bool  `json:"blocking"`
}

// MediaCommentsPage holds one page of comments plus cursors for the same media.
type MediaCommentsPage struct {
	Comments      []Comment `json:"comments"`
	HasMore       bool      `json:"has_more"`
	NextMaxID     string    `json:"next_max_id"`
	NextMinID     string    `json:"next_min_id"`
	CommentCount  int       `json:"comment_count"`
}

// VideoUploadMeta is required metadata for direct video upload when file probing is not used.
type VideoUploadMeta struct {
	WidthPx      int
	HeightPx     int
	DurationMS   int
}

// Hashtag is a tag search result with optional media count and icon URL.
type Hashtag struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	MediaCount    int    `json:"media_count"`
	ProfilePicURL string `json:"profile_pic_url"`
}

// Track is a music search result entry.
type Track struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URI   string `json:"uri"`
}
