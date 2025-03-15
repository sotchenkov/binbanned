package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sotchenkov/binbanned/internal/config"
	"github.com/sotchenkov/binbanned/internal/logprocessor"

	"github.com/fsnotify/fsnotify"
	"github.com/hpcloud/tail"
)

// Start initiates watching of the log directory and tails log files.
func Start(ctx context.Context, cfg *config.Config) {
	var wg sync.WaitGroup
	activeFiles := make(map[string]struct{})
	filesMutex := sync.Mutex{}

	// startTail starts tailing the specified file if not already active.
	startTail := func(filePath string) {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			log.Printf("Error getting absolute path: %v", err)
			return
		}
		filesMutex.Lock()
		if _, exists := activeFiles[absPath]; exists {
			filesMutex.Unlock()
			return
		}
		activeFiles[absPath] = struct{}{}
		filesMutex.Unlock()

		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			tailFile(ctx, path, cfg.ParseAll)
		}(absPath)
	}

	// Tail all existing files in the directory.
	files, err := os.ReadDir(cfg.LogDir)
	if err != nil {
		log.Printf("Error reading directory %s: %v", cfg.LogDir, err)
	} else {
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			startTail(filepath.Join(cfg.LogDir, file.Name()))
		}
	}

	// Set up fsnotify watcher for new files.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Error creating fsnotify watcher: %v", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(cfg.LogDir)
	if err != nil {
		log.Printf("Error adding directory %s to watcher: %v", cfg.LogDir, err)
		return
	}

	// Listen for file creation events.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					// Start tailing the new file.
					startTail(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	// Wait until context cancellation.
	<-ctx.Done()
	// Wait for all tail goroutines to finish.
	wg.Wait()
}

// tailFile tails a file and processes each new line.
func tailFile(ctx context.Context, filePath string, parseAll bool) {
	tailConfig := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Poll:      true,
	}
	if parseAll {
		tailConfig.Location = &tail.SeekInfo{Whence: 0} // Start from beginning
	} else {
		tailConfig.Location = &tail.SeekInfo{Whence: 2} // Start from end (io.SeekEnd)
	}

	t, err := tail.TailFile(filePath, tailConfig)
	if err != nil {
		log.Printf("Error tailing file %s: %v", filePath, err)
		return
	}
	defer t.Cleanup()

	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-t.Lines:
			if !ok {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			// Pass the current file name to the processor.
			logprocessor.ProcessLine(line.Text, filePath)
		}
	}
}
