package fetcher

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type CacheManager struct {
	cacheDir string
	maxAge   time.Duration
}

func NewCacheManager(maxAge time.Duration) (*CacheManager, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(cacheDir, "cetatenie-analyzer")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &CacheManager{
		cacheDir: dir,
		maxAge:   maxAge,
	}, nil
}

func (cm *CacheManager) GetFilePath(year int) string {
	return filepath.Join(cm.cacheDir, "anul_"+strconv.Itoa(year)+".pdf")
}

func (cm *CacheManager) Cleanup() error {
	files, err := os.ReadDir(cm.cacheDir)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > cm.maxAge {
			filePath := filepath.Join(cm.cacheDir, file.Name())
			_ = os.Remove(filePath)
		}
	}
	return nil
}

func (cm *CacheManager) FileExists(year int) bool {
	filePath := cm.GetFilePath(year)
	_, err := os.Stat(filePath)
	return err == nil
}
