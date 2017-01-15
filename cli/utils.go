package cli

import "strings"

var knownVideoFileExtensions = []string{
	"webm",
	"mkv",
	"flv",
	"vob",
	"ogg",
	"ogv",
	"drc",
	"avi",
	"qt",
	"wmv",
	"rm",
	"rmvb",
	"asf",
	"amv",
	"mp4",
	"m4p",
	"m4v",
	"mpg",
	"mp2",
	"mpeg",
	"mpe",
	"mpv",
	"m2v",
	"3gp",
	"3g2",
	"mxf",
	"f4v",
	"f4p",
	"f4a",
	"f4b",
}

var videoFileBlacklist = []string{
	"sample", "RAGB",
}

func isKnownVideoFileExtension(extension string) bool {
	for _, ext := range knownVideoFileExtensions {
		if ext == extension {
			return true
		}
	}
	return false
}

func isAllowedVideoFile(name string) bool {
	for _, word := range videoFileBlacklist {
		if strings.Contains(strings.ToLower(name), word) {
			return false
		}
	}
	return true
}

func extractFileName(path string) string {
	fileName := path
	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		fileName = parts[len(parts)-1]
	}
	return fileName
}
