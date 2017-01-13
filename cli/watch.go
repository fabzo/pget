package cli

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"pget/premiumize"
	"pget/watcher"
	"strings"
	"time"
)

const premiumizeFinishedStatus = "finished"
const torrentListCheckInterval = 60 * time.Second
const boltDBFile = "pget.db"
const torrentsBucket = "torrents"

func (c *Cli) openBoltDB() error {
	c.boltMutex.Lock()
	defer c.boltMutex.Unlock()

	if c.bolt == nil {
		db, err := bolt.Open(boltDBFile, 0600, nil)
		if err != nil {
			return err
		}
		c.bolt = db
	}
	return nil
}

func (c *Cli) WatchAndUpload(directory string, strict bool, deleteAfterUpload bool) {
	stat, err := os.Stat(directory)
	if err != nil {
		fmt.Printf("Unable to retrieve directory stats: %s\n", err.Error())
		return
	}

	if !stat.IsDir() {
		fmt.Printf("%s is not a directory\n", directory)
		return
	}

	if err := c.openBoltDB(); err != nil {
		fmt.Printf("Unable to open database for upload/download tracking: %s\n", err.Error())
		return
	}
	c.bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(torrentsBucket))
		return err
	})

	watcher := watcher.New(watcher.FileWatcherConfig{
		BaseDir:      directory,
		MatchPattern: ".*?\\.torrent",
		ScanInterval: 5 * time.Second,
	})

	pathCh := make(chan string)
	watcher.AddWatcher(pathCh)
	watcher.Run()

	var path string
	for {
		path = <-pathCh
		c.processTorrentFile(directory, path, strict, deleteAfterUpload)
	}
}

func (c *Cli) setupWatcher(watcher *fsnotify.Watcher, directory string) error {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		err = watcher.Add(path)
		if err != nil {
			fmt.Printf("Failed to add directory %s to watcher: %s\n", path, err.Error())
		}
		fmt.Printf("Watching %s\n", path)
		return nil
	})

	return err
}

func (c *Cli) processTorrentFile(basePath string, filePath string, strict bool, deleteAfterUpload bool) {
	location := extractLocation(basePath, filePath)

	// TODO: Check if torrent is already in our db
	// TODO: Check torrent file integrity

	resp, err := c.upload(filePath)

	if err != nil {
		fmt.Printf("Failed to upload %s: %s\n", filePath, err.Error())
	} else {
		err := os.Remove(filePath)
		if err != nil {
			fmt.Printf("Could not delete torrent file after processing: %s\n", err.Error())
		}
		err = c.bolt.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(torrentsBucket))
			err := bucket.Put([]byte(resp.ID), []byte(location))
			return err
		})
		if err != nil {
			fmt.Printf("Failed to store torrent %s in database, this torrent will not be automatically downloaded: %s\n", filePath, err.Error())
		}
	}
}

func (c *Cli) upload(filePath string) (premiumize.UploadResponse, error) {
	if strings.HasSuffix(filePath, ".torrent") {
		return c.premiumize.UploadTorrentFile(filePath)
	} else {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return premiumize.UploadResponse{}, err
		}
		return c.premiumize.UploadMagnetLink(string(content))
	}
	return premiumize.UploadResponse{}, nil
}

func extractLocation(basePath string, filePath string) string {
	path := filePath[len(basePath)+1:]

	if strings.ContainsRune(path, os.PathSeparator) {
		file := filepath.Base(path)
		return filePath[:len(filePath)-len(file)-1]
	}
	return ""
}

func (c *Cli) WatchAndDownload(targetDirectory string, videoOnly bool, flatten bool, strict bool, deleteDownloaded bool, createSyncFile bool) {
	if strict {
		if err := c.openBoltDB(); err != nil {
			fmt.Printf("Unable to open database for upload/download tracking: %s\n", err.Error())
			return
		}
	}

	done := make(chan bool)
	go func() {
		for {
			if createSyncFile {
				c.createSyncFile(targetDirectory)
			}

			torrents, err := c.premiumize.ListTorrents()
			if err != nil {
				fmt.Printf("Could not retrieve list of torrents: %s\n", err.Error())
			} else {
				for _, transfer := range torrents.Transfers {
					isFinished := c.isTorrentFinished(transfer.Status)
					hasBeenUploaded := c.hasBeenUploadedWhenStrict(strict, transfer)

					if isFinished && hasBeenUploaded {
						id, err := c.DownloadTorrent(transfer.Name, targetDirectory, videoOnly, flatten, "")
						if deleteDownloaded && err == nil {
							c.premiumize.DeleteTorrent(id)
						}
					}
				}
			}
			if createSyncFile {
				c.deleteSyncFile(targetDirectory)
			}
			time.Sleep(torrentListCheckInterval)
		}
	}()

	<-done
}

func (c *Cli) createSyncFile(directory string) {
	file, err := os.OpenFile(path.Join(directory, ".sync"), os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Could not create sync file: %v", err)
	}
	file.Close()
}

func (c *Cli) deleteSyncFile(directory string) {
	err := os.Remove(path.Join(directory, ".sync"))
	if err != nil {
		fmt.Printf("Could not delete sync file: %v", err)
	}
}

func (c *Cli) isTorrentFinished(status string) bool {
	return status == premiumizeFinishedStatus
}

func (c *Cli) hasBeenUploadedWhenStrict(strict bool, transfer premiumize.TorrentItem) bool {
	if !strict {
		return true
	}

	err := c.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(torrentsBucket))
		name := bucket.Get([]byte(transfer.ID))
		if name == nil {
			return fmt.Errorf("Not found")
		}
		return nil
	})

	return err == nil
}
