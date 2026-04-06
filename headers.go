package ig

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	igcfg "github.com/felipeinf/instago/config"
)

func (c *Client) buildBaseHeaders() http.Header {
	locale := c.locale
	if locale == "" {
		locale = "en_US"
	}
	igLocale := locale
	acceptLang := []string{"en-US"}
	if locale != "" {
		lang := locale
		for i := 0; i < len(lang); i++ {
			if lang[i] == '_' {
				lang = lang[:i] + "-" + lang[i+1:]
				break
			}
		}
		if lang != "en-US" {
			acceptLang = append([]string{lang}, acceptLang...)
		}
	}
	acceptJoined := acceptLang[0]
	for i := 1; i < len(acceptLang); i++ {
		acceptJoined += ", " + acceptLang[i]
	}
	r := c.rng
	bwKbps := float64(r.Intn(500000)+2500000) / 1000
	bwB := r.Intn(85000000) + 5000000
	bwMs := r.Intn(7000) + 2000
	h := http.Header{}
	h.Set("X-IG-App-Locale", igLocale)
	h.Set("X-IG-Device-Locale", igLocale)
	h.Set("X-IG-Mapped-Locale", igLocale)
	h.Set("X-Pigeon-Session-Id", c.generateUUID("UFS-", "-1"))
	h.Set("X-Pigeon-Rawclienttime", fmt.Sprintf("%.3f", float64(time.Now().UnixNano())/1e9))
	h.Set("X-IG-Bandwidth-Speed-KBPS", strconv.FormatFloat(bwKbps, 'f', 3, 64))
	h.Set("X-IG-Bandwidth-TotalBytes-B", strconv.Itoa(bwB))
	h.Set("X-IG-Bandwidth-TotalTime-MS", strconv.Itoa(bwMs))
	h.Set("X-IG-App-Startup-Country", c.country)
	h.Set("X-Bloks-Version-Id", c.bloksVersioningID)
	h.Set("X-IG-WWW-Claim", "0")
	h.Set("X-Bloks-Is-Layout-RTL", "false")
	h.Set("X-Bloks-Is-Panorama-Enabled", "true")
	h.Set("X-IG-Device-ID", c.uuids.UUID)
	h.Set("X-IG-Family-Device-ID", c.uuids.PhoneID)
	h.Set("X-IG-Android-ID", c.uuids.AndroidDeviceID)
	h.Set("X-IG-Timezone-Offset", strconv.Itoa(c.timezoneOffset))
	h.Set("X-IG-Connection-Type", "WIFI")
	h.Set("X-IG-Capabilities", "3brTv10=")
	h.Set("X-IG-App-ID", c.appID)
	h.Set("Priority", "u=3")
	h.Set("User-Agent", c.userAgent)
	h.Set("Accept-Language", acceptJoined)
	h.Set("X-MID", c.mid)
	h.Set("Host", igcfg.APIDomain)
	h.Set("X-FB-HTTP-Engine", "Liger")
	h.Set("Connection", "keep-alive")
	h.Set("X-FB-Client-IP", "True")
	h.Set("X-FB-Server-Cluster", "True")
	uid := c.userID()
	h.Set("IG-INTENDED-USER-ID", strconv.FormatInt(uid, 10))
	h.Set("X-IG-Nav-Chain", "9MV:self_profile:2,ProfileMediaTabFragment:self_profile:3,9Xf:self_following:4")
	h.Set("X-IG-SALT-IDS", strconv.Itoa(r.Intn(100000)+1061162222))
	if uid != 0 {
		nextYear := float64(time.Now().Unix()) + 31536000
		h.Set("IG-U-DS-USER-ID", strconv.FormatInt(uid, 10))
		h.Set("IG-U-IG-DIRECT-REGION-HINT", fmt.Sprintf("LLA,%d,%.0f:01f7bae7d8b131877d8e0ae1493252280d72f6d0d554447cb1dc9049b6b2c507c08605b7", uid, nextYear))
		h.Set("IG-U-SHBID", fmt.Sprintf("12695,%d,%.0f:01f778d9c9f7546cf3722578fbf9b85143cd6e5132723e5c93f40f55ca0459c8ef8a0d9f", uid, nextYear))
		h.Set("IG-U-SHBTS", fmt.Sprintf("%d,%d,%.0f:01f7ace11925d038808007d0282b75b8059844855da27e23c90a362270fddfb3fae7e28", time.Now().Unix(), uid, nextYear))
		h.Set("IG-U-RUR", fmt.Sprintf("RVA,%d,%.0f:01f7f627f9ae4ce2874b2e04463efdb184340968b1b006fa88cb4cc69a942a04201e544c", uid, nextYear))
	}
	if c.igURur != "" {
		h.Set("IG-U-RUR", c.igURur)
	}
	if c.igWWWClaim != "" {
		h.Set("X-IG-WWW-Claim", c.igWWWClaim)
	}
	return h
}
