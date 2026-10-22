package analyzer

import (
	"ghostops/internal/model"
	"sort"
	"strings"
	"time"
)

type Rule struct {
	Name        string
	Description string
	Match       func(entry *model.LogEntry) bool
}

func DefaultRules() []Rule {
	return []Rule{
		{
			Name:        "database-pool-exhausted",
			Description: "Database connection pool exhausted and crashed system",
			Match: func(e *model.LogEntry) bool {
				return e.Level == model.LevelFatal && strings.Contains(e.Message, "connection pool exhausted")
			},
		},
		{
			Name:        "database-connection-failure",
			Description: "Service failed to connect to database",
			Match: func(e *model.LogEntry) bool {
				return (e.Level == model.LevelError || e.Level == model.LevelFatal) &&
					strings.Contains(e.Message, "Failed to connect to database")
			},
		},
		{
			Name:        "payment-gateway-failure",
			Description: "Payment gateway error or timeout",
			Match: func(e *model.LogEntry) bool {
				return (e.Level == model.LevelError || e.Level == model.LevelFatal) &&
					strings.Contains(e.Message, "Payment gateway")
			},
		},
		{
			Name:        "application-exception",
			Description: "Unhandled application exception in service controller",
			Match: func(e *model.LogEntry) bool {
				return (e.Level == model.LevelError || e.Level == model.LevelFatal) &&
					strings.Contains(e.Message, "NullPointerException")
			},
		},
		{
			Name:        "high-memory-warning",
			Description: "High memory utilization threshold reached",
			Match: func(e *model.LogEntry) bool {
				return (e.Level == model.LevelWarn || e.Level == model.LevelError) &&
					strings.Contains(e.Message, "High memory usage detected")
			},
		},
		{
			Name:        "database-pool-warning",
			Description: "Database connection pool near capacity threshold",
			Match: func(e *model.LogEntry) bool {
				return e.Level == model.LevelWarn &&
					strings.Contains(e.Message, "Database connection pool reaching threshold")
			},
		},
		{
			Name:        "rate-limit-warning",
			Description: "Client approaching rate limit threshold",
			Match: func(e *model.LogEntry) bool {
				return e.Level == model.LevelWarn &&
					strings.Contains(e.Message, "Rate limit approaching")
			},
		},
	}
}

type Analyzer struct {
	rules     []Rule
	stats     model.Stats
	incidents map[string]*model.IncidentGroup 
}

func New(rules ...Rule) *Analyzer {
	if len(rules) == 0 {
		rules = DefaultRules()
	}
	return &Analyzer{
		rules: rules,
		stats: model.Stats{
			ByLevel:   make(map[string]int),
			ByService: make(map[string]int),
		},
		incidents: make(map[string]*model.IncidentGroup),
	}
}

func (a *Analyzer) RecordMalformed() {
	a.stats.TotalMalformed++
}

func (a *Analyzer) Process(entry *model.LogEntry) {
	if entry == nil {
		a.RecordMalformed()
		return
	}

	a.stats.TotalValid++
	a.stats.ByLevel[entry.Level]++
	a.stats.ByService[entry.Service]++

	for _, rule := range a.rules {
		if rule.Match(entry) {
			a.recordIncident(rule.Name, entry)
		}
	}
}

func (a *Analyzer) recordIncident(ruleName string, entry *model.LogEntry) {
	key := ruleName + "\x00" + entry.Service
	group, exists := a.incidents[key]
	if !exists {
		group = &model.IncidentGroup{
			RuleName:  ruleName,
			Service:   entry.Service,
			TopLevel:  entry.Level,
			Count:     1,
			FirstSeen: entry.Timestamp,
			LastSeen:  entry.Timestamp,
		}
		a.incidents[key] = group
		return
	}

	group.Count++
	if entry.Timestamp.Before(group.FirstSeen) {
		group.FirstSeen = entry.Timestamp
	}
	if entry.Timestamp.After(group.LastSeen) {
		group.LastSeen = entry.Timestamp
	}
	if model.LevelRank[entry.Level] > model.LevelRank[group.TopLevel] {
		group.TopLevel = entry.Level
	}
}

func (a *Analyzer) Report(logFile string) model.Report {
	incidentList := make([]model.IncidentGroup, 0, len(a.incidents))
	for _, group := range a.incidents {
		incidentList = append(incidentList, *group)
	}

	sort.Slice(incidentList, func(i, j int) bool {
		if incidentList[i].Count != incidentList[j].Count {
			return incidentList[i].Count > incidentList[j].Count 
		}
		if incidentList[i].RuleName != incidentList[j].RuleName {
			return incidentList[i].RuleName < incidentList[j].RuleName
		}
		return incidentList[i].Service < incidentList[j].Service
	})

	return model.Report{
		GeneratedAt: time.Now().UTC(),
		LogFile:     logFile,
		Stats:       a.stats,
		Incidents:   incidentList,
	}
}
