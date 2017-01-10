package main

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"pget/cli"
	"pget/premiumize"
	"sync"
)

func main() {
	viper.AutomaticEnv()
	viper.SetConfigName("pget")
	viper.AddConfigPath("/etc/pget/")
	viper.AddConfigPath("$HOME/.pget")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s\n", err))
	}

	customerID := viper.GetString("customer_id")
	pin := viper.GetString("pin")

	if customerID == "" || pin == "" {
		fmt.Println("Customer ID or pin not configured")
		return
	}

	premiumizeClient := premiumize.New(customerID, pin)

	application := kingpin.New("pget", "Premiumize Get")
	debugFlag := application.Flag("debug", "Dump parsed premiumize.me responses").Bool()

	listCommand := application.Command("list", "List torrents")

	treeCommand := application.Command("tree", "Print tree of the torrent files")
	torrentNameArg := treeCommand.Arg("name", "Name of the torrent").String()

	downloadCommand := application.Command("download", "Downloads the content of a given torrent")
	downloadNameArg := downloadCommand.Arg("name", "Name of the torrent").String()
	downloadVideoOnlyFlag := downloadCommand.Flag("video-only", "Only download video files (also ignores samples)").Short('v').Bool()
	downloadFlattenFlag := downloadCommand.Flag("flatten", "Ignore directories").Short('f').Bool()
	downloadStopAfterFlag := downloadCommand.Flag("stop-after", "Stop download after x [43mb, 4gb]").Short('s').String()
	downloadDirectoryFlag := downloadCommand.Flag("directory", "Directory to which the files should be downloaded").Short('d').Default(".").String()

	uploadCommand := application.Command("upload", "Upload a torrent file or magnet link")
	uploadLink := uploadCommand.Arg("link", "Torrent file or magnet link").String()

	watchCommand := application.Command("watch", "Watch for local or remote files to upload/download")

	watchUploadFlag := watchCommand.Flag("upload", "Directory to watch for new torrent files to upload").Default("-").String()
	watchDeleteFlag := watchCommand.Flag("delete", "Delete torrent file after upload").Bool()

	watchDownloadFlag := watchCommand.Flag("download", "Directory to which torrents are downloaded").Default("-").String()
	watchStrictDownloadFlag := watchCommand.Flag("strict", "Only download torrents that have also been uploaded by this tool").Bool()
	watchVideoOnlyFlag := watchCommand.Flag("video-only", "Only download video files (also ignores samples)").Short('v').Bool()
	watchFlattenFlag := watchCommand.Flag("flatten", "Ignore directories").Short('f').Bool()

	cli := cli.New(premiumizeClient)

	switch kingpin.MustParse(application.Parse(os.Args[1:])) {

	case listCommand.FullCommand():
		premiumizeClient.SetDebug(*debugFlag)
		cli.ListTorrents()

	case treeCommand.FullCommand():
		premiumizeClient.SetDebug(*debugFlag)
		cli.TreeTorrents(*torrentNameArg)

	case downloadCommand.FullCommand():
		premiumizeClient.SetDebug(*debugFlag)
		cli.DownloadTorrent(*downloadNameArg, *downloadDirectoryFlag, *downloadVideoOnlyFlag, *downloadFlattenFlag, *downloadStopAfterFlag)

	case uploadCommand.FullCommand():
		premiumizeClient.SetDebug(*debugFlag)
		cli.Upload(*uploadLink)

	case watchCommand.FullCommand():
		premiumizeClient.SetDebug(*debugFlag)

		var wg sync.WaitGroup

		if *watchUploadFlag != "-" {
			wg.Add(1)
			go func() {
				cli.WatchAndUpload(*watchUploadFlag, *watchStrictDownloadFlag, *watchDeleteFlag)
				wg.Done()
			}()
		}

		if *watchDownloadFlag != "-" {
			wg.Add(1)
			go func() {
				cli.WatchAndDownload(*watchDownloadFlag, *watchVideoOnlyFlag, *watchFlattenFlag, *watchStrictDownloadFlag)
				wg.Done()
			}()
		}

		wg.Wait()
	}
}
