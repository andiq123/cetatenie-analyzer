package parser

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/ledongthuc/pdf"
)

type PDFParser interface {
	ReadPdf(data io.ReadSeeker, search string) (FindState, error)
	GetYear(search string) (int, error)
}

type pdfParser struct{}

func New() PDFParser {
	return &pdfParser{}
}

type FindState int

const (
	StateNotFound FindState = iota
	StateFoundButNotResolved
	StateFoundAndResolved
)

var stateMessages = map[FindState]string{
	StateNotFound:            "negăsit",
	StateFoundButNotResolved: "găsit dar nerezolvat",
	StateFoundAndResolved:    "găsit și rezolvat",
}

func GetStateMessage(state FindState) string {
	if msg, ok := stateMessages[state]; ok {
		return msg
	}
	return "Unknown state"
}

func (p *pdfParser) ReadPdf(data io.ReadSeeker, search string) (FindState, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, data)
	if err != nil {
		return StateNotFound, fmt.Errorf("error reading PDF data: %v", err)
	}

	reader, err := pdf.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return StateNotFound, fmt.Errorf("error creating PDF reader: %v", err)
	}

	numPages := reader.NumPage()
	if numPages == 0 {
		return StateNotFound, nil
	}

	// Channel to receive results from goroutines
	resultChan := make(chan FindState, numPages)
	// Channel to limit concurrent goroutines (optional)
	semaphore := make(chan struct{}, 10) // Limit to 10 concurrent goroutines

	var wg sync.WaitGroup

	for i := 1; i <= numPages; i++ {
		wg.Add(1)
		go func(pageNum int) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			page := reader.Page(pageNum)
			if page.V.IsNull() {
				return
			}

			text, err := page.GetPlainText(nil)
			if err != nil {
				// You might want to handle errors differently in concurrent context
				return
			}

			index := strings.Index(text, search)
			if index != -1 {
				const offset = 43
				end := min(index+offset, len(text))
				if strings.Contains(text[index:end], "/P/") {
					resultChan <- StateFoundAndResolved
					return
				}
				resultChan <- StateFoundButNotResolved
				return
			}
		}(i)
	}

	// Close the result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results as they come in
	for result := range resultChan {
		if result == StateFoundAndResolved {
			// Return immediately if we find the best possible match
			return result, nil
		}
		// For StateFoundButNotResolved, keep checking other pages
		// in case we find a StateFoundAndResolved later
	}

	return StateNotFound, nil
}

// GetYear extracts the year from a decree number string
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
