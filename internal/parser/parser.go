package parser

import (
	"bufio"
	"fmt"
	"ghostops/internal/model"
	"io"
	"strings"
	"time"
)

type Scanner struct {
	s         *bufio.Scanner
	entry     *model.LogEntry
	malformed bool
	rawLine   string
	err       error
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{s: bufio.NewScanner(r)}
}

func (sc *Scanner) Scan() bool {
	for sc.s.Scan() {
		line := sc.s.Text()
		if strings.TrimSpace(line) == "" {
			continue 
		}
		sc.rawLine = line
		entry, err := parseLine(line)
		if err != nil {
			sc.entry = nil
			sc.malformed = true
		} else {
			sc.entry = entry
			sc.malformed = false
		}
		return true
	}
	sc.err = sc.s.Err() 
	return false
}

func (sc *Scanner) Entry() *model.LogEntry { return sc.entry }

func (sc *Scanner) IsMalformed() bool { return sc.malformed }

func (sc *Scanner) RawLine() string { return sc.rawLine }

func (sc *Scanner) Err() error { return sc.err }

func parseLine(line string) (*model.LogEntry, error) {
	parts := strings.SplitN(line, " ", 4)
	if len(parts) < 4 {
		return nil, fmt.Errorf("expected 4 fields, got %d", len(parts))
	}

	ts, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp %q", parts[0])
	}

	level := parts[1]
	if !model.ValidLevels[level] {
		return nil, fmt.Errorf("unknown level %q", level)
	}

	return &model.LogEntry{
		Timestamp: ts,
		Level:     level,
		Service:   parts[2],
		Message:   parts[3],
	}, nil
}
