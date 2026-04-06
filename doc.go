// Package ig is an unofficial Go client for Instagram private and public HTTP APIs.
//
// Import path: github.com/felipeinf/instago. The declared package name is ig (for example ig.NewClient).
// This project is not affiliated with Meta or Instagram; you are responsible for your use and for complying with their terms and applicable law.
//
// The symbol index below pkg.go.dev is alphabetical. This overview follows a business flow: sign-in, configure the client, call features by domain, then models and low-level hooks.
//
// # Getting started
//
// NewClient, then either Login (username, password, optional 2FA code) or LoadSettings from a file saved by DumpSettings.
//
// # Authentication and session
//
// PreLoginFlow — optional warm-up before Login.
//
// Login — password sign-in; pass the app/SMS code when TwoFactorRequired.
//
// LoginFlow — lightweight post-login feed calls (reels tray + timeline).
//
// GetReelsTrayFeed, GetTimelineFeed — feed endpoints used after login or for pagination (max_id).
//
// Logout — end server session.
//
// LoadSettings, DumpSettings — persist and restore full client state on *Client.
//
// LoadSettingsFromFile, DumpSettingsToFile — read/write Settings JSON without a live Client.
//
// # Client configuration
//
// SetLogger, SetProxy, SetLocale, SetTimezoneOffset, SetDeviceSettings, SetUserAgent, SetUUIDs.
//
// # Logged-in account
//
// AccountInfo — current user from accounts/current_user.
//
// # Users and profiles
//
// UserIDFromUsername, UserInfoByUsername (cache flag), UserInfo — resolved profiles.
//
// UserInfoByUsernameGQL, UserInfoByUsernameV1 — lower-level paths used internally or for debugging.
//
// SearchUsersV1, SearchUsersFB — user search variants.
//
// # Media and files
//
// UserMedias — high level: GraphQL timeline with REST fallback.
//
// UserMediasWithSleep, UserMediasGQL, UserMediasV1 — explicit backends and pacing (sleepSec, cursors).
//
// UserMediasPaginatedGQL, UserMediasPaginatedV1 — single-page fetches (end_cursor / next_max_id). Media may include CarouselMedia (REST carousel_media or GQL sidecar children).
//
// RankToken — token for user feed pagination.
//
// MediaInfoV1, PhotoDownload, PhotoDownloadByURL — single media metadata and photo download.
//
// UserClipsPaginatedV1, UserClipsV1 — reels (clips/user/) with enriched Media (counts, usertags, caption).
//
// DownloadToFile, StoryDownloadByURL, StoryDownload — generic and story-aware downloads (see DownloadOptions).
//
// # Stories
//
// UserStories — GraphQL first, then REST story reel.
//
// UserStoriesV1 — private API only; items include reel_mentions, story_hashtags, story_locations when present.
//
// StoryInfo — resolve one item by story PK string.
//
// # Direct messages
//
// DirectInboxChunk, DirectPendingChunk, DirectSpamChunk — list threads (default one request: limit=20, thread_message_limit=10 for inbox).
//
// DirectThreadPage — one GET for thread messages (default limit=20; exact count is server-defined).
//
// DirectSendText — broadcast text or link.
//
// DirectSendPhoto, DirectSendVideo — rupload then configure_photo/configure_video (video requires VideoUploadMeta).
//
// DirectPhotoRupload, DirectVideoRupload — low-level rupload steps; DirectMessageMediaURL — best URL for received media.
//
// # Comments
//
// MediaCommentsFirstPage — first comments page (page size is server-defined).
//
// MediaCommentsFetch, MediaCommentsNext — pagination via next_max_id / next_min_id (has_more_comments / has_more_headload_comments).
//
// # Friendship
//
// FriendshipWith — friendships/show/ (following, followed_by, requests, private, etc.).
//
// Follow, Unfollow — friendships/create and friendships/destroy (signed action data).
//
// MutualFriendsPage — friendships/{id}/mutual_friends/ (rank_token; optional max_id for next page).
//
// UserFollowersPage — friendships/{id}/followers/ (rank_token, search_surface; optional max_id).
//
// # Search and discovery
//
// FbsearchTopsearchFlat, FbsearchRecent, FbsearchSuggestedProfiles, FbsearchPlaces.
//
// SearchHashtags, SearchMusic.
//
// # Public and web helpers
//
// PublicRequest, PublicGraphqlRequest — unauthenticated HTTP/GraphQL (rate limited internally).
//
// WebProfileInfo — web_profile_info JSON for a username.
//
// FetchPasswordEncryptionKeys — same keys as package password for custom login flows.
//
// # Instagram challenges
//
// ChallengeGET, ChallengePOST — thin wrappers over PrivateRequest for challenge URLs.
//
// # Low-level private API
//
// PrivateRequest with PrivateRequestOpts — escape hatch for any signed or unsigned endpoint.
//
// ParseAuthorizationHeader — decode Bearer IGT authorization payloads.
//
// # Models (structs)
//
// All in types.go, in this order: UUIDs, DeviceSettings, Settings (session snapshot); UserShort, User, Account; Media, Comment, Story (mentions/hashtags/locations), DirectThread, DirectMessage, Friendship, MediaCommentsPage, VideoUploadMeta; Hashtag, Track.
//
// Request and client options: PrivateRequestOpts in client.go; DownloadOptions in download.go; Logger in client.go.
//
// # Errors and debugging
//
// Errors from Instagram are typed in github.com/felipeinf/instago/igerrors — use errors.As (e.g. LoginRequired, TwoFactorRequired, UserNotFound).
//
// Client.LastJSON — last decoded private API JSON. Client.LastHTTPResponse — last HTTP response when set.
//
// Client.OverrideAppVersion — before LoadSettings, allow replacing stored app build with a supported profile from config.
//
// # Subpackages
//
//   - github.com/felipeinf/instago/config — API URLs, device defaults, app version profiles
//   - github.com/felipeinf/instago/encoding — JSON compaction and signed_body helpers (not the standard library encoding tree)
//   - github.com/felipeinf/instago/password — password encryption for Login
//   - github.com/felipeinf/instago/igerrors — HTTP/JSON error mapping and helpers
package ig
