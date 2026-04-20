# Types

Defined in [`types.go`](../types.go) unless noted. JSON tags match Instagram payloads.

---

## Session & device

**`UUIDs`** — `PhoneID`, `UUID`, `ClientSessionID`, `AdvertisingID`, `AndroidDeviceID`, `RequestID`, `TraySessionID`.

**`DeviceSettings`** — Android build: `AndroidVersion`, `AndroidRelease`, `DPI`, `Resolution`, `Manufacturer`, `Device`, `Model`, `CPU`, optional app/version/bloks fields.

**`Settings`** — On-disk snapshot: `UUIDs`, `Mid`, `IgURur`, `IgWWWClaim`, `AuthorizationData`, `Cookies`, `LastLogin`, `DeviceSettings`, `UserAgent`, `Country`, `CountryCode`, `Locale`, `TimezoneOffset`.

---

## Users

**`UserShort`** — `PK`, `Username`, `FullName`, `ProfilePicURL`, `IsPrivate`, `IsVerified`.

**`User`** — embeds `UserShort`; adds `Biography`, `ExternalURL`, `IsBusiness`, `FollowerCount`, `FollowingCount`, `MediaCount`, `ProfilePicURLHD`, `PublicEmail`, `ContactPhone`.

**`Account`** — Current user: `PK`, `Username`, `FullName`, `Biography`, `ExternalURL`, `Email`, `PhoneNumber`.

**`Usertag`** — `User`, `X`, `Y` (position).

---

## Media & comments

**`Media`** — `PK`, `ID`, `Code`, `MediaType`, `ProductType`, `TakenAt`, `ThumbnailURL`, `VideoURL`, `CaptionText`, `User`, `LikeCount`, `CommentCount`, `Usertags`, `PlayCount`, `CarouselMedia`.

**`Comment`** — `PK`, `Text`, `User`, `CreatedAt`, `LikeCount`.

**`MediaCommentsPage`** — `Comments`, `HasMore`, `NextMaxID`, `NextMinID`, `CommentCount`.

---

## Stories

**`Story`** — `PK`, `ID`, `Code`, `MediaType`, `TakenAt`, `ThumbnailURL`, `VideoURL`, `ProductType`, `User`, `Mentions`, `Hashtags`, `Locations`.

**`StoryMention`**, **`StoryHashtag`**, **`StoryLocation`** — Stickers (user / hashtag / place + geometry).

---

## Direct messages

**`DirectThread`** — `PK`, `ID`, `Users`, `Messages`, `Title`, `ThreadType`, `Inviter`, `LastActivity`.

**`DirectMessage`** — `ID`, `ThreadID`, `UserID`, `TimestampUS`, `Text`, `ItemType`, media URL fields.

---

## Social graph & search

**`Friendship`** — `UserID`, `Following`, `FollowedBy`, `IncomingRequest`, `OutgoingRequest`, `IsPrivate`, `IsRestricted`, `Blocking`.

**`Hashtag`** — `ID`, `Name`, `MediaCount`, `ProfilePicURL`.

**`Track`** — Music search: `ID`, `Title`, `URI`.

---

## Options & uploads

**`DirectInboxOptions`** ([`direct.go`](../direct.go)) — `SelectedFilter`, `Box` (`"primary"` / `"general"`), `ThreadMessageLimit`, `Limit`, `Cursor`.

**`DirectThreadOptions`** — `Cursor`, `Limit`.

**`VideoUploadMeta`** — `WidthPx`, `HeightPx`, `DurationMS` (DM video).

**`PrivateRequestOpts`** ([`client.go`](../client.go)) — `Endpoint`, `Data`, `Params`, `Login`, `WithSignature`, `ExtraHeaders`, `ExtraSig`, `Domain`, `RawBody`, `IsRawJSONBody`.

**`DownloadOptions`** ([`download.go`](../download.go)) — `Timeout`.

**`Logger`** ([`client.go`](../client.go)) — `Infof`, `Debugf`, `Warnf`.
