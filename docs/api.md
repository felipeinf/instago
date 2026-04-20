# API & examples

Import: `import ig "github.com/felipeinf/instago"`. Create clients with `ig.NewClient()`.

---

## Session & authentication

| Method | Notes |
|--------|--------|
| `NewClient()` | New client with defaults |
| `PreLoginFlow()` | Optional warm-up before login |
| `Login(user, pass, verificationCode)` | Empty code if no 2FA |
| `LoginFlow()` | Light post-login calls |
| `GetReelsTrayFeed(reason)` / `GetTimelineFeed(reason, maxID)` | Feeds, pagination via `maxID` |
| `Logout()` | End session |
| `LoadSettings(path, overrideAppVersion)` / `DumpSettings(path)` | Persist session on `Client` |
| `LoadSettingsFromFile` / `DumpSettingsToFile` | Package-level JSON without a live client |

```go
c := ig.NewClient()
if err := c.Login("user", "pass", ""); err != nil { /* handle */ }
_ = c.DumpSettings("session.json")

c2 := ig.NewClient()
_ = c2.LoadSettings("session.json", false)
```

---

## Client configuration

`SetLogger`, `SetProxy`, `SetLocale`, `SetTimezoneOffset`, `SetDeviceSettings`, `SetUserAgent`, `SetUUIDs`.

---

## Account

| Method | Notes |
|--------|--------|
| `AccountInfo()` | Logged-in user (`Account`) |

```go
a, err := c.AccountInfo()
```

---

## Users & profiles

| Method | Notes |
|--------|--------|
| `UserIDFromUsername(username)` | Resolve PK |
| `UserInfoByUsername(username, useCache)` | Main high-level profile |
| `UserInfo(userID, useCache)` | By numeric id |
| `UserInfoByUsernameGQL` / `UserInfoByUsernameV1` | Lower-level paths |
| `SearchUsersV1` / `SearchUsersFB` | User search |

```go
u, err := c.UserInfoByUsername("instagram", true)
```

---

## Media, feed, reels, download

| Method | Notes |
|--------|--------|
| `UserMedias` | High level (GQL + fallback) |
| `UserMediasWithSleep` / `UserMediasGQL` / `UserMediasV1` | Explicit backend / pacing |
| `UserMediasPaginatedGQL` / `UserMediasPaginatedV1` | Single page + cursor |
| `RankToken()` | Pagination helper for some feeds |
| `MediaInfoV1` | One post by PK |
| `PhotoDownload` / `PhotoDownloadByURL` | Save photo to disk |
| `UserClipsV1` / `UserClipsPaginatedV1` | Reels (clips) |
| `DownloadToFile` | Generic HTTP GET to file |
| `StoryDownload` / `StoryDownloadByURL` | Story assets |

```go
items, err := c.UserMedias(userID, 12)
m, err := c.MediaInfoV1(mediaPK)
path, err := c.PhotoDownload(mediaPK, "out", "folder")
```

---

## Stories

| Method | Notes |
|--------|--------|
| `UserStories` / `UserStoriesV1` | Story reels for a user |
| `StoryInfo(storyPK)` | One story item |

---

## Direct messages

| Method | Notes |
|--------|--------|
| `DirectInboxChunk(opt)` | Inbox page; see `DirectInboxOptions` |
| `DirectPendingChunk` / `DirectSpamChunk` | Pending / spam |
| `DirectThreadPage(threadID, opt)` | Messages in a thread |
| `DirectSendText` / `DirectSendPhoto` / `DirectSendVideo` | Outgoing |
| `DirectPhotoRupload` / `DirectVideoRupload` | Low-level upload steps |
| `DirectMessageMediaURL(m)` | **Package function** — pick URL from a `DirectMessage` |

Defaults: inbox ~20 threads, ~10 preview messages per thread; thread page limit ~20 (server-defined).

```go
threads, next, err := c.DirectInboxChunk(ig.DirectInboxOptions{})
full, _, err := c.DirectThreadPage(threadID, ig.DirectThreadOptions{})
```

---

## Comments

| Method | Notes |
|--------|--------|
| `MediaCommentsFirstPage` | First page |
| `MediaCommentsFetch` / `MediaCommentsNext` | Pagination with min/max ids |

---

## Friendship

| Method | Notes |
|--------|--------|
| `FriendshipWith` | Status flags |
| `Follow` / `Unfollow` | Mutate relationship |
| `MutualFriendsPage` | Mutual followers |
| `UserFollowersPage` | Followers list + cursor |

---

## Search & discovery

`FbsearchTopsearchFlat`, `FbsearchRecent`, `FbsearchSuggestedProfiles`, `FbsearchPlaces`, `SearchHashtags`, `SearchMusic`.

---

## Public / web (limited auth)

| Method | Notes |
|--------|--------|
| `PublicRequest` / `PublicGraphqlRequest` | Unauthenticated HTTP / GraphQL |
| `WebProfileInfo` | `web_profile_info` JSON |
| `FetchPasswordEncryptionKeys` | For custom login flows |

---

## Challenges & low-level

| Method | Notes |
|--------|--------|
| `ChallengeGET` / `ChallengePOST` | Challenge URL flows |
| `PrivateRequest(opts)` | Any private endpoint; see `PrivateRequestOpts` |

---

## Debugging

- `Client.LastJSON` — last decoded private JSON  
- `Client.LastHTTPResponse` — last HTTP response when set  
- `Client.OverrideAppVersion` — set before `LoadSettings` to replace stored app build  

Use [`igerrors`](https://pkg.go.dev/github.com/felipeinf/instago/igerrors) with `errors.As` (e.g. `TwoFactorRequired`, `LoginRequired`).
