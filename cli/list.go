package cli

import (
	"fmt"
	"github.com/dustin/go-humanize"
)

func (c *Cli) ListTorrents() {
	list, err := c.premiumize.ListTorrents()
	if err != nil {
		panic(err)
	}

	for _, transfer := range list.Transfers {
		fmt.Printf("* %s [%s] [%s]\n", transfer.Name, transfer.Status, humanize.Bytes(uint64(transfer.Size)))
	}
}
