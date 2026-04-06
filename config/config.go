package config

const (
	// APIDomain is the hostname for private Instagram API requests.
	APIDomain = "i.instagram.com"
)

// UserAgentBase is the Android Instagram User-Agent template; the client replaces braced placeholders.
const UserAgentBase = "Instagram {app_version} " +
	"Android ({android_version}/{android_release}; " +
	"{dpi}; {resolution}; {manufacturer}; " +
	"{model}; {device}; {cpu}; {locale}; {version_code})"

// DeviceDefaults holds default emulated device fields (OS, screen, SoC) before app version is set.
type DeviceDefaults struct {
	AndroidVersion int
	AndroidRelease string
	DPI            string
	Resolution     string
	Manufacturer   string
	Device         string
	Model          string
	CPU            string
}

// DefaultDevice is the default handset profile used when constructing a new client.
var DefaultDevice = DeviceDefaults{
	AndroidVersion: 26,
	AndroidRelease: "8.0.0",
	DPI:            "480dpi",
	Resolution:     "1080x1920",
	Manufacturer:   "OnePlus",
	Device:         "devitron",
	Model:          "6T Dev",
	CPU:            "qcom",
}

// AppProfile identifies one Instagram Android app build (semantic version, Play version code, bloks id).
type AppProfile struct {
	AppVersion        string
	VersionCode       string
	BloksVersioningID string
}

// AppProfiles maps known app version strings to signing and header metadata.
var AppProfiles = map[string]AppProfile{
	"364.0.0.35.86": {
		AppVersion:        "364.0.0.35.86",
		VersionCode:       "374010953",
		BloksVersioningID: "8ccf54aad76788a6ca03ddfc33afcdcf692f2f5a3ba814ea73d5facba7fa2c2d",
	},
	"385.0.0.47.74": {
		AppVersion:        "385.0.0.47.74",
		VersionCode:       "378906843",
		BloksVersioningID: "a8973d49a9cc6a6f65a4997c10216ce2a06f65a517010e64885e92029bb19221",
	},
}

// AppProfileList returns all entries from AppProfiles in an unspecified order.
func AppProfileList() []AppProfile {
	out := make([]AppProfile, 0, len(AppProfiles))
	for _, v := range AppProfiles {
		out = append(out, v)
	}
	return out
}

// LoginExperiments is the comma-separated Android experiment flags string used in login-related payloads.
const LoginExperiments = "ig_android_reg_nux_headers_cleanup_universe," +
	"ig_android_device_detection_info_upload," +
	"ig_android_nux_add_email_device," +
	"ig_android_gmail_oauth_in_reg," +
	"ig_android_device_info_foreground_reporting," +
	"ig_android_device_verification_fb_signup," +
	"ig_android_direct_main_tab_universe_v2," +
	"ig_android_passwordless_account_password_creation_universe," +
	"ig_android_direct_add_direct_to_android_native_photo_share_sheet," +
	"ig_growth_android_profile_pic_prefill_with_fb_pic_2," +
	"ig_account_identity_logged_out_signals_global_holdout_universe," +
	"ig_android_quickcapture_keep_screen_on," +
	"ig_android_device_based_country_verification," +
	"ig_android_login_identifier_fuzzy_match," +
	"ig_android_reg_modularization_universe," +
	"ig_android_security_intent_switchoff," +
	"ig_android_device_verification_separate_endpoint," +
	"ig_android_suma_landing_page," +
	"ig_android_sim_info_upload," +
	"ig_android_smartlock_hints_universe," +
	"ig_android_fb_account_linking_sampling_freq_universe," +
	"ig_android_retry_create_account_universe," +
	"ig_android_caption_typeahead_fix_on_o_universe"

// SupportedCapabilities is the capability map list sent with some private feed requests.
var SupportedCapabilities = []map[string]string{
	{
		"value": "119.0,120.0,121.0,122.0,123.0,124.0,125.0,126.0,127.0,128.0," +
			"129.0,130.0,131.0,132.0,133.0,134.0,135.0,136.0,137.0,138.0," +
			"139.0,140.0,141.0,142.0",
		"name": "SUPPORTED_SDK_VERSIONS",
	},
	{"value": "14", "name": "FACE_TRACKER_VERSION"},
	{"value": "ETC2_COMPRESSION", "name": "COMPRESSION"},
	{"value": "gyroscope_enabled", "name": "gyroscope"},
}

const (
	// PublicWebURL is the Instagram website origin (trailing slash).
	PublicWebURL = "https://www.instagram.com/"
	// GraphQLPublicAPIURL is the public GraphQL endpoint base URL (query string appended by callers).
	GraphQLPublicAPIURL = "https://www.instagram.com/graphql/query/"
)

// DefaultIGAppID is the default Instagram web app id (x-ig-app-id) for logged-out web API calls.
const DefaultIGAppID = "936619743392459"

// FBAnalyticsAppID is the Facebook analytics application id carried on some client payloads.
const FBAnalyticsAppID = "567067343352427"
