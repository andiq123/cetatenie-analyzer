package decree

import (
	"fmt"

	"github.com/andiq123/cetatenie-analyzer/internal/fetcher"
	"github.com/andiq123/cetatenie-analyzer/internal/timer"
)

// Processor defines the interface for processing decree searches
type Processor interface {
	Handle(search string) (FindState, *timer.TimeReport, error)
	CleanUpCache() error
}

type service struct {
	fetcher fetcher.FileFetcher
	parser  IParser
}

func NewProcessor() Processor {
	f, err := fetcher.New()
	if err != nil {
		panic(fmt.Errorf("failed to create fetcher: %v", err))
	}

	return &service{
		fetcher: f,
		parser:  newParser(),
	}
}

func (s *service) Handle(search string) (FindState, *timer.TimeReport, error) {
	year, err := s.parser.GetYear(search)
	if err != nil {
		return StateNotFound, &timer.TimeReport{}, fmt.Errorf("format dosar invalid: %v", err)
	}

	fetchTimer := timer.NewTimer()
	fetchTimer.Start()
	dataBytes, err := s.fetcher.GetFile(year)
	if err != nil {
		return StateNotFound, &timer.TimeReport{}, fmt.Errorf("nu am putut obține fișierul pentru anul %d: %v", year, err)
	}
	fetchTimer.Stop()
	fetchTime := fetchTimer.Duration()

	parseTimer := timer.NewTimer()
	parseTimer.Start()
	state, err := s.parser.ReadPdf(dataBytes, search)
	if err != nil {
		return StateNotFound, &timer.TimeReport{}, fmt.Errorf("eroare la analiza documentului: %v", err)
	}
	parseTimer.Stop()
	parseTime := parseTimer.Duration()

	timeReport := timer.NewTimeReport(fetchTime, parseTime)

	return state, timeReport, nil
}

func (s *service) CleanUpCache() error {
	return s.fetcher.CleanUpCache()
}
