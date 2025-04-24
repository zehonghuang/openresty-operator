package watcher

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func DiscoverWatchDirs(root string) ([]string, error) {
	seen := make(map[string]struct{})
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !strings.HasSuffix(path, ".conf") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 {
			dir := filepath.Dir(path)
			seen[dir] = struct{}{}
			fmt.Println("[agent] will watch directory:", dir)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0, len(seen))
	for d := range seen {
		dirs = append(dirs, d)
	}
	return dirs, nil
}
