package ig

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadOptions configures HTTP timeouts for media downloads.
type DownloadOptions struct {
	// Timeout is the HTTP client timeout for the GET request.
	Timeout time.Duration
}

// DownloadToFile streams rawURL to destPath using the client's User-Agent; opt defaults to a 60s timeout when nil.
func (c *Client) DownloadToFile(rawURL, destPath string, opt *DownloadOptions) error {
	if opt == nil {
		opt = &DownloadOptions{Timeout: 60 * time.Second}
	}
	client := &http.Client{Timeout: opt.Timeout}
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("download: http %d", resp.StatusCode)
	}
	dir := filepath.Dir(destPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

// StoryDownloadByURL downloads media from a direct CDN URL into folder; filename may include or omit an extension.
func (c *Client) StoryDownloadByURL(mediaURL, filename, folder string) (string, error) {
	u, err := url.Parse(mediaURL)
	if err != nil {
		return "", err
	}
	fname := filepath.Base(u.Path)
	if fname == "" || fname == "." {
		return "", fmt.Errorf("download: url has no filename path")
	}
	ext := ""
	if i := strings.LastIndex(fname, "."); i >= 0 {
		ext = fname[i:]
		fname = fname[:i]
	}
	outName := fname + ext
	if filename != "" {
		if strings.Contains(filename, ".") {
			outName = filename
		} else {
			outName = filename + ext
		}
	}
	dir := folder
	if dir == "" {
		dir = "."
	}
	dest := filepath.Join(dir, outName)
	if err := c.DownloadToFile(mediaURL, dest, nil); err != nil {
		return "", err
	}
	return dest, nil
}

// StoryDownload resolves storyPK via StoryInfo and downloads thumbnail or video to folder.
func (c *Client) StoryDownload(storyPK, filename, folder string) (string, error) {
	st, err := c.StoryInfo(storyPK)
	if err != nil {
		return "", err
	}
	u := st.ThumbnailURL
	if st.MediaType == 2 {
		u = st.VideoURL
	}
	return c.StoryDownloadByURL(u, filename, folder)
}
