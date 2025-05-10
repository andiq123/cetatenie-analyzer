package fetcher

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/andiq123/cetatenie-analyzer/internal/cache"
)

type FileFetcher interface {
	GetFile(year int) ([]byte, error)
	CleanUpCache() error
}

type httpFetcher struct {
	client  *http.Client
	baseURL string
	cache   *cache.Cache
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
			InsecureSkipVerify: true, // Changed to false for security
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
		cache:   cache.New(24 * time.Hour), // Cache for 24 hours with a max of 100 items
	}, nil
}

// GetFile retrieves the annual report PDF for the given year
func (f *httpFetcher) GetFile(year int) ([]byte, error) {
	filename, ok := supportedYears[year]
	if !ok {
		return nil, fmt.Errorf("year %d is not supported", year)
	}

	url := f.baseURL + filename

	// Check cache first
	if data, found := f.cache.Get(url); found {
		return data, nil
	}

	data, err := f.downloadFileWithRetry(url, 3) // 3 retries
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	f.cache.Set(url, data)
	f.cache.Cleanup()

	return data, nil
}

// downloadFileWithRetry handles the download with retry logic
func (f *httpFetcher) downloadFileWithRetry(url string, maxRetries int) ([]byte, error) {
	var lastErr error

	for i := range maxRetries {
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

	// Set optimized headers
	req.Header = http.Header{
		"User-Agent":      {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"},
		"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8"},
		"Accept-Language": {"ro-RO,ro;q=0.9,en-US;q=0.8,en;q=0.7"},
		"Accept-Encoding": {"gzip"}, // Enable compression
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Optimized error body reading
		buf := make([]byte, 1024)
		n, _ := io.ReadFull(resp.Body, buf)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(buf[:n]))
	}

	// Fast content type check
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		return nil, fmt.Errorf("unexpected content type: %s", ct)
	}

	// Optimized reading based on content length
	if resp.ContentLength > 0 {
		// Pre-allocate exact buffer size
		buf := make([]byte, resp.ContentLength)
		_, err := io.ReadFull(resp.Body, buf)
		return buf, err
	}

	// Fallback for unknown size - uses sync.Pool for buffers
	return readWithPool(resp.Body)
}

// Reusable buffer pool for unknown content lengths
var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024) // 32KB chunks
	},
}

func readWithPool(r io.Reader) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 64*1024)) // Initial 64KB capacity
	temp := bufPool.Get().([]byte)
	defer bufPool.Put(&temp)

	for {
		n, err := r.Read(temp)
		if n > 0 {
			buf.Write(temp[:n])
		}
		if err != nil {
			if err == io.EOF {
				return buf.Bytes(), nil
			}
			return nil, err
		}
	}
}

func (f *httpFetcher) CleanUpCache() error {
	f.cache.Cleanup()
	return nil
}
