package decree_processor

import (
	"fmt"
	"path/filepath"

	"github.com/andiq123/cetatenie-analyzer/internal/fetcher"
	"github.com/andiq123/cetatenie-analyzer/internal/parser"
)

// Processor defines the interface for processing decree searches
type Processor interface {
	Handle(search string) (parser.FindState, error)
}

type service struct {
	fetcher fetcher.FileFetcher
	parser  parser.PDFParser
}

// New creates a new decree processor service
func New() Processor {
	f, err := fetcher.New()
	if err != nil {
		panic(fmt.Errorf("failed to create fetcher: %v", err))
	}

	return &service{
		fetcher: f,
		parser:  parser.New(),
	}
}

// Handle processes a decree search request
func (s *service) Handle(search string) (parser.FindState, error) {
	year, err := s.parser.GetYear(search)
	if err != nil {
		return parser.StateNotFound, fmt.Errorf("format dosar invalid: %v", err)
	}

	tempPath, err := s.fetcher.GetFile(year)
	if err != nil {
		return parser.StateNotFound, fmt.Errorf("nu am putut obține fișierul pentru anul %d: %v", year, err)
	}

	state, err := s.parser.ReadPdf(tempPath, search)
	if err != nil {
		return parser.StateNotFound, fmt.Errorf("eroare la analiza documentului: %v", err)
	}

	return state, nil
}

func (s *service) cleanupTempFile(path string) {
	if path != "" && filepath.Ext(path) == ".pdf" {
		// In production: os.Remove(path)
	}
}
