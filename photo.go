package ig

import (
	"fmt"
)

// PhotoDownloadByURL downloads a file from mediaURL into folder (same implementation as StoryDownloadByURL).
func (c *Client) PhotoDownloadByURL(mediaURL, filename, folder string) (string, error) {
	return c.StoryDownloadByURL(mediaURL, filename, folder)
}

// MediaInfoV1 returns metadata for a single media id via media/{id}/info/.
func (c *Client) MediaInfoV1(mediaPK int64) (Media, error) {
	res, err := c.PrivateRequest(PrivateRequestOpts{
		Endpoint:      fmt.Sprintf("media/%d/info/", mediaPK),
		WithSignature: true,
	})
	if err != nil {
		return Media{}, err
	}
	items, ok := res["items"].([]any)
	if !ok || len(items) == 0 {
		return Media{}, fmt.Errorf("ig: media info missing items")
	}
	m, _ := items[0].(map[string]any)
	return extractMediaV1(m), nil
}

// PhotoDownload fetches MediaInfoV1 and saves the photo thumbnail when media_type is photo (1).
func (c *Client) PhotoDownload(mediaPK int64, filename, folder string) (string, error) {
	med, err := c.MediaInfoV1(mediaPK)
	if err != nil {
		return "", err
	}
	if med.MediaType != 1 {
		return "", fmt.Errorf("ig: media is not a photo")
	}
	if filename == "" {
		filename = fmt.Sprintf("%s_%d", med.User.Username, mediaPK)
	}
	return c.PhotoDownloadByURL(med.ThumbnailURL, filename, folder)
}
