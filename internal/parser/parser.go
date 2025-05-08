package parser

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

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
	// Read all data into a bytes.Reader which implements both io.ReaderAt and io.Seeker
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, data)
	if err != nil {
		return StateNotFound, fmt.Errorf("error reading PDF data: %v", err)
	}

	reader, err := pdf.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
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
			const offset = 43
			end := min(index+offset, len(text))
			if strings.Contains(text[index:end], "/P/") {
				return StateFoundAndResolved, nil
			}
			return StateFoundButNotResolved, nil
		}
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
