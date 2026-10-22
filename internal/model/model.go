package model

import "time"

const (
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
)

var ValidLevels = map[string]bool{
	LevelInfo:  true,
	LevelWarn:  true,
	LevelError: true,
	LevelFatal: true,
}

var LevelRank = map[string]int{
	LevelInfo:  0,
	LevelWarn:  1,
	LevelError: 2,
	LevelFatal: 3,
}

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Service   string
	Message   string
}

type Stats struct {
	TotalValid     int            `json:"total_valid"`
	TotalMalformed int            `json:"total_malformed"`
	ByLevel        map[string]int `json:"by_level"`
	ByService      map[string]int `json:"by_service"`
}

type IncidentGroup struct {
	RuleName  string    `json:"rule"`
	Service   string    `json:"service"`
	TopLevel  string    `json:"severity"` 
	Count     int       `json:"count"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

type Report struct {
	GeneratedAt time.Time       `json:"generated_at"`
	LogFile     string          `json:"log_file"`
	Stats       Stats           `json:"stats"`
	Incidents   []IncidentGroup `json:"incidents"`
}
