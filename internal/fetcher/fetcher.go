package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type FileFetcher interface {
	GetFile(year int) (string, error)
	CleanupCache() error
}

type httpFetcher struct {
	client  *http.Client
	baseURL string
	cache   *CacheManager
}

var supportedYears = map[int]string{
	2020: "art_11_anul_2020.pdf",
	2021: "art_11_anul_2021.pdf",
	2022: "art_11_anul_2022.pdf",
	2023: "art_11_anul_2023.pdf",
	2024: "art_11_anul_2024.pdf",
	2025: "art_11_anul_2025.pdf",
}

// New creates a new file fetcher instance
func New(maxCacheAge time.Duration) (FileFetcher, error) {
	cache, err := NewCacheManager(maxCacheAge)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache manager: %v", err)
	}

	return &httpFetcher{
		client:  &http.Client{},
		baseURL: "https://cetatenie.just.ro/storage/2023/11/",
		cache:   cache,
	}, nil
}

// GetFile retrieves the annual report PDF for the given year
func (f *httpFetcher) GetFile(year int) (string, error) {
	filename, ok := supportedYears[year]
	if !ok {
		return "", fmt.Errorf("anul %d nu este suportat", year)
	}

	filePath := f.cache.GetFilePath(year)

	// Return cached file if exists
	if f.cache.FileExists(year) {
		return filePath, nil
	}

	// Download the file if not cached
	url := f.baseURL + filename
	if err := f.downloadFile(url, filePath); err != nil {
		return "", fmt.Errorf("eroare la descărcare: %v", err)
	}

	return filePath, nil
}

func (f *httpFetcher) CleanupCache() error {
	return f.cache.Cleanup()
}

func (f *httpFetcher) downloadFile(url, filePath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cod de stare HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
