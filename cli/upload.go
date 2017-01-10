package cli

import (
	"strings"
)

func (c *Cli) Upload(link string) {
	if strings.HasPrefix(link, "magnet") {
		c.premiumize.UploadMagnetLink(link)
	} else {
		c.premiumize.UploadTorrentFile(link)
	}
}
