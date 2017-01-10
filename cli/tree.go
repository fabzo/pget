package cli

import (
	"bytes"
	"fmt"
	"pget/premiumize"
)

func (c *Cli) TreeTorrents(name string) {
	torrentInfo, err := c.premiumize.FindTorrentByName(name)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	torrent, err := c.premiumize.BrowseTorrent(torrentInfo.Hash)

	fmt.Println(".")
	printTorrent(0, torrent.Content)
}

func printTorrent(level int, torrent map[string]premiumize.TorrentContent) {
	index := 0
	for key, value := range torrent {
		index++
		fmt.Printf("%s %s\n", treePrefix(level, index == len(torrent)), key)
		if value.Children != nil {
			printTorrent(level+1, value.Children)
		}
	}
}

func treePrefix(level int, isLast bool) string {
	var buffer bytes.Buffer
	for i := 0; i < level; i++ {
		buffer.WriteString("│   ")
	}
	if isLast {
		buffer.WriteString("└──")
	} else {
		buffer.WriteString("├──")
	}
	return buffer.String()
}
