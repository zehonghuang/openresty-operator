package watcher

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

func WatchDirectory(path string, onChange func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(path); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", path, err)
	}

	fmt.Printf("[reload-agent] watching directory: %s\n", path)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			fmt.Printf("[reload-agent] change detected: %s\n", event)
			onChange()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintln(os.Stderr, "[reload-agent] fsnotify error:", err)
		}
	}
}
