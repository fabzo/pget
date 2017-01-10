package watcher

import (
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

type FileWatcher struct {
	config       FileWatcherConfig
	watcher      []chan<- string
	watcherMutex sync.RWMutex
	running      bool
	runningMutex sync.Mutex
}

type FileWatcherConfig struct {
	BaseDir      string
	MatchPattern string
	ScanInterval time.Duration
}

func New(config FileWatcherConfig) *FileWatcher {
	return &FileWatcher{
		config: config,
	}
}

func (w *FileWatcher) AddWatcher(watcher chan<- string) {
	w.watcherMutex.Lock()
	w.watcher = append(w.watcher, watcher)
	w.watcherMutex.Unlock()
}

func (w *FileWatcher) Run() {
	w.runningMutex.Lock()
	if w.running {
		w.runningMutex.Unlock()
		return
	}
	w.running = true
	w.runningMutex.Unlock()

	go w.watchDirectory()
}

func (w *FileWatcher) watchDirectory() {
	match := regexp.MustCompile(w.config.MatchPattern)
	for {
		filepath.Walk(w.config.BaseDir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			if match.MatchString(path) {
				w.watcherMutex.RLock()
				for _, watcher := range w.watcher {
					watcher <- path
				}
				w.watcherMutex.RUnlock()
			}

			return nil
		})
		time.Sleep(w.config.ScanInterval)
	}
}
