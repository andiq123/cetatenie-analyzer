package decree

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/ledongthuc/pdf"
)

const OFFSET = 43

type IParser interface {
	ReadPdf(data []byte, search string) (FindState, error)
	GetYear(search string) (int, error)
}

type pdfParser struct{}

func newParser() IParser {
	return &pdfParser{}
}

func (p *pdfParser) ReadPdf(data []byte, search string) (FindState, error) {
	// Create reader directly from the byte slice
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return StateNotFound, fmt.Errorf("error creating PDF reader: %v", err)
	}

	// Process pages concurrently with a worker pool
	numPages := reader.NumPage()
	if numPages == 0 {
		return StateNotFound, nil
	}

	// Use a reasonable number of workers based on available CPUs
	numWorkers := min(runtime.NumCPU(), numPages)

	type pageResult struct {
		state FindState
		err   error
	}

	jobs := make(chan int, numPages)
	results := make(chan pageResult, numPages)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case pageNum, ok := <-jobs:
					if !ok {
						return
					}

					page := reader.Page(pageNum)
					if page.V.IsNull() {
						results <- pageResult{StateNotFound, nil}
						continue
					}

					text, err := page.GetPlainText(nil)
					if err != nil {
						results <- pageResult{StateNotFound, fmt.Errorf("error reading page %d: %v", pageNum, err)}
						continue
					}

					index := strings.Index(text, search)
					if index != -1 {
						end := min(index+OFFSET, len(text))
						if strings.Contains(text[index:end], "/P/") {
							results <- pageResult{StateFoundAndResolved, nil}
						} else {
							results <- pageResult{StateFoundButNotResolved, nil}
						}
					} else {
						results <- pageResult{StateNotFound, nil}
					}
				}
			}
		}()
	}

	go func() {
		for i := 1; i <= numPages; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.err != nil {
			cancel()
			return StateNotFound, result.err
		}

		if result.state != StateNotFound {
			cancel()
			return result.state, nil
		}
	}

	return StateNotFound, nil
}

func (p *pdfParser) GetYear(search string) (int, error) {
	parts := strings.Split(search, "/")
	if len(parts) != 3 || parts[1] != "RD" {
		return 0, fmt.Errorf("format invalid, folosește [număr]/RD/[an]")
	}

	yearStr := parts[2]
	if len(yearStr) != 4 {
		return 0, fmt.Errorf("anul trebuie să aibă 4 cifre")
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return 0, fmt.Errorf("an invalid: %s", yearStr)
	}

	if year < 2000 || year > 2100 {
		return 0, fmt.Errorf("anul %d este în afara intervalului valid", year)
	}

	return year, nil
}
