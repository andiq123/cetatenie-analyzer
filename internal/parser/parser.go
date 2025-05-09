package parser

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
)

const OFFSET = 43

type PDFParser interface {
	ReadPdf(data []byte, search string) (FindState, error)
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

func (p *pdfParser) ReadPdf(data []byte, search string) (FindState, error) {
	// Create reader directly from the byte slice
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return StateNotFound, fmt.Errorf("error creating PDF reader: %v", err)
	}

	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}

		text, err := page.GetPlainText(nil)
		if err != nil {
			return StateNotFound, fmt.Errorf("error reading page %d: %v", i, err)
		}

		index := strings.Index(text, search)
		if index != -1 {
			end := min(index+OFFSET, len(text))
			if strings.Contains(text[index:end], "/P/") {
				return StateFoundAndResolved, nil
			}
			return StateFoundButNotResolved, nil
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
