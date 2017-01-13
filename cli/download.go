package cli

import (
	"fmt"
	"github.com/cavaliercoder/grab"
	"github.com/dustin/go-humanize"
	"os"
	"path/filepath"
	"pget/premiumize"
	"sort"
	"strings"
	"time"
)

const typeFile = "file"
const typeDir = "dir"

type DownloadTask struct {
	Destination string
	URL         string
	Size        uint64
}

type DownloadTaskSorter []DownloadTask

func (a DownloadTaskSorter) Len() int      { return len(a) }
func (a DownloadTaskSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DownloadTaskSorter) Less(i, j int) bool {
	return strings.Compare(a[i].Destination, a[j].Destination) < 0
}

func (c *Cli) DownloadTorrent(name string, targetDirectory string, videoOnly bool, flatten bool, stopAfter string) (string, error) {
	var bytes uint64 = 0
	if stopAfter != "" {
		var err error
		bytes, err = humanize.ParseBytes(stopAfter)
		if err != nil {
			fmt.Printf("Unable to parse %s. Error: %s\n", stopAfter, err.Error())
		}
	}

	torrentInfo, err := c.premiumize.FindTorrentByName(name)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	torrent, err := c.premiumize.BrowseTorrent(torrentInfo.Hash)

	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	tasks := createDownloadList(targetDirectory, torrent.Content, videoOnly, flatten)
	return torrentInfo.ID, download(tasks, bytes)
}

func download(tasks []DownloadTask, stopAfterBytes uint64) error {
	sort.Sort(DownloadTaskSorter(tasks))

	var totalBytes uint64
	for _, task := range tasks {
		if stopAfterBytes != 0 {
			if _, err := os.Stat(task.Destination); err != nil {
				totalBytes += task.Size
				if totalBytes > stopAfterBytes {
					fmt.Printf("Stopping download. Reached %s. The next download would overstep the %s limit.", humanize.Bytes(totalBytes-task.Size), humanize.Bytes(stopAfterBytes))
					return nil
				}
			}
		}

		return downloadTask(task)
	}
	return nil
}

func createDownloadList(root string, torrent map[string]premiumize.TorrentContent, videoOnly bool, flatten bool) []DownloadTask {
	var downloadList []DownloadTask

	for _, value := range torrent {
		if value.Type == typeFile {
			if videoOnly && !isVideo(value) {
				continue
			}

			downloadList = append(downloadList, toDownloadTask(root, value, flatten))
		} else {
			tasks := createDownloadList(root, value.Children, videoOnly, flatten)
			downloadList = append(downloadList, tasks...)
		}
	}
	return downloadList
}

func isVideo(torrentFile premiumize.TorrentContent) bool {
	if isKnownVideoFileExtension(torrentFile.Ext) && isAllowedVideoFile(torrentFile.Name) {
		return true
	}
	return false
}

func toDownloadTask(root string, torrentFile premiumize.TorrentContent, flatten bool) DownloadTask {
	path := torrentFile.Path
	fileName := extractFileName(path)

	destination := root + string(filepath.Separator) + fileName
	if !flatten {
		destination = root + string(filepath.Separator) + path
	}

	return DownloadTask{
		Destination: destination,
		URL:         torrentFile.URL,
		Size:        uint64(torrentFile.Size),
	}
}

func downloadTask(task DownloadTask) error {
	respCh, err := grab.GetAsync(task.Destination, task.URL)
	if err != nil {
		fmt.Printf("Error downloading %s: %s\n", task.Destination, err.Error())
		return err
	}

	resp := <-respCh
	fmt.Println("")
	for !resp.IsComplete() {
		fmt.Printf("\033[1A   %s [%s / %s] (%d%%)\033[K\n", task.Destination, humanize.Bytes(resp.BytesTransferred()), humanize.Bytes(resp.Size), int(100*resp.Progress()))
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Printf("\033[1A\033[K")
	if resp.Error != nil {
		fmt.Printf("   Error downloading %s: %v\n", task.URL, resp.Error)
		return err
	}
	fmt.Printf("   %s [%s]\n", task.Destination, humanize.Bytes(resp.Size))
	return nil
}
