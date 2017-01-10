package cli

import (
	"github.com/boltdb/bolt"
	"pget/premiumize"
	"sync"
)

type Cli struct {
	premiumize *premiumize.Client
	bolt       *bolt.DB
	boltMutex  sync.Mutex
}

func New(client *premiumize.Client) *Cli {
	return &Cli{
		premiumize: client,
	}
}
