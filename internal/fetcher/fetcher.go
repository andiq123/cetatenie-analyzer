package fetcher

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

type FileFetcher interface {
	GetFile(year int) ([]byte, error)
}

type httpFetcher struct {
	client  *http.Client
	baseURL string
}

var supportedYears = map[int]string{
	2020: "art_11_anul_2020.pdf",
	2021: "art_11_anul_2021.pdf",
	2022: "art_11_anul_2022.pdf",
	2023: "art_11_anul_2023.pdf",
	2024: "art_11_anul_2024.pdf",
	2025: "art_11_anul_2025.pdf",
}

// New creates a new HTTP file fetcher instance with proper configuration
func New() (FileFetcher, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false, // Changed to false for security
		},
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
		DisableKeepAlives:  false,
	}

	return &httpFetcher{
		client: &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second, // Added timeout
		},
		baseURL: "https://cetatenie.just.ro/storage/2023/11/",
	}, nil
}

// GetFile retrieves the annual report PDF for the given year
func (f *httpFetcher) GetFile(year int) ([]byte, error) {
	filename, ok := supportedYears[year]
	if !ok {
		return nil, fmt.Errorf("year %d is not supported", year)
	}

	url := f.baseURL + filename
	data, err := f.downloadFileWithRetry(url, 3) // 3 retries
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return data, nil
}

// downloadFileWithRetry handles the download with retry logic
func (f *httpFetcher) downloadFileWithRetry(url string, maxRetries int) ([]byte, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(time.Second * time.Duration(i*i)) // Exponential backoff
		}

		data, err := f.downloadFile(url)
		if err == nil {
			return data, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("after %d attempts: %w", maxRetries, lastErr)
}

// downloadFile handles a single download attempt
func (f *httpFetcher) downloadFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	// Set realistic headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ro-RO,ro;q=0.9,en-US;q=0.8,en;q=0.7")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // Read partial response for error details
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Verify content type if you expect PDF
	if ct := resp.Header.Get("Content-Type"); ct != "application/pdf" {
		return nil, fmt.Errorf("unexpected content type: %s", ct)
	}

	return io.ReadAll(resp.Body)
}
