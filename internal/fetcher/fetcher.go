package fetcher

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
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

// New creates a new in-memory file fetcher instance
func New() (FileFetcher, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Skip TLS verification (not recommended for production)
		},
	}
	return &httpFetcher{
		client: &http.Client{
			Transport: transport,
		},
		baseURL: "https://cetatenie.just.ro/storage/2023/11/",
	}, nil
}

// GetFile retrieves the annual report PDF for the given year and returns it as an io.ReadSeeker
func (f *httpFetcher) GetFile(year int) ([]byte, error) {
	filename, ok := supportedYears[year]
	if !ok {
		return nil, fmt.Errorf("anul %d nu este suportat", year)
	}

	url := f.baseURL + filename
	data, err := f.downloadFile(url)
	if err != nil {
		return nil, fmt.Errorf("eroare la descărcare: %v", err)
	}

	return data, nil
}

func (f *httpFetcher) downloadFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
