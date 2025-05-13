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

const (
	OFFSET = 43
	// Buffer size for text extraction
	textBufferSize = 1024 * 1024 // 1MB
	// Maximum number of concurrent workers
	maxWorkers = 8
	// Batch size for page processing
	pageBatchSize = 10
)

type IParser interface {
	ReadPdf(data []byte, search string) (FindState, error)
	GetYear(search string) (int, error)
}

type pdfParser struct {
	// Buffer pool for text extraction
	bufferPool sync.Pool
}

type pageResult struct {
	state FindState
	err   error
}

func newParser() IParser {
	return &pdfParser{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, textBufferSize))
			},
		},
	}
}

func (p *pdfParser) ReadPdf(data []byte, search string) (FindState, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return StateNotFound, fmt.Errorf("error creating PDF reader: %v", err)
	}

	numPages := reader.NumPage()
	if numPages == 0 {
		return StateNotFound, nil
	}

	// Calculate optimal number of workers
	numWorkers := min(min(maxWorkers, runtime.NumCPU()), numPages)

	// Create channels for job distribution and results
	jobs := make(chan []int, (numPages+pageBatchSize-1)/pageBatchSize)
	results := make(chan pageResult, numWorkers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go p.worker(ctx, &wg, reader, search, jobs, results)
	}

	// Distribute work in batches
	go func() {
		for i := 1; i <= numPages; i += pageBatchSize {
			end := min(i+pageBatchSize-1, numPages)
			batch := make([]int, 0, pageBatchSize)
			for j := i; j <= end; j++ {
				batch = append(batch, j)
			}
			select {
			case <-ctx.Done():
				return
			case jobs <- batch:
			}
		}
		close(jobs)
	}()

	// Wait for completion and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
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

func (p *pdfParser) worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	reader *pdf.Reader,
	search string,
	jobs <-chan []int,
	results chan<- pageResult,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case batch, ok := <-jobs:
			if !ok {
				return
			}

			for _, pageNum := range batch {
				select {
				case <-ctx.Done():
					return
				default:
					result := p.processPage(reader, pageNum, search)
					if result.state != StateNotFound || result.err != nil {
						results <- result
						return
					}
				}
			}
		}
	}
}

func (p *pdfParser) processPage(reader *pdf.Reader, pageNum int, search string) pageResult {
	page := reader.Page(pageNum)
	if page.V.IsNull() {
		return pageResult{StateNotFound, nil}
	}

	// Get buffer from pool
	buf := p.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		p.bufferPool.Put(buf)
	}()

	// Extract text with buffer reuse
	text, err := page.GetPlainText(nil)
	if err != nil {
		return pageResult{StateNotFound, fmt.Errorf("error reading page %d: %v", pageNum, err)}
	}

	// Optimize search by checking if the search string exists before detailed analysis
	if !strings.Contains(text, search) {
		return pageResult{StateNotFound, nil}
	}

	// Perform detailed search
	index := strings.Index(text, search)
	if index != -1 {
		end := min(index+OFFSET, len(text))
		if strings.Contains(text[index:end], "/P/") {
			return pageResult{StateFoundAndResolved, nil}
		}
		return pageResult{StateFoundButNotResolved, nil}
	}

	return pageResult{StateNotFound, nil}
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
