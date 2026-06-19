package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
)

type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Message   string         `json:"message"`
	Source    string         `json:"source"`
	Fields    map[string]any `json:"fields,omitempty"`
}

var sources = []string{
	"api-gateway", "app-server", "auth-service", "db-primary",
	"cache-redis", "worker-queue", "search-indexer", "filebeat",
	"kibana", "elasticsearch", "load-balancer", "monitoring",
}

var levels = []string{"ERROR", "WARN", "INFO", "DEBUG", "TRACE"}

var messages = map[string][]string{
	"ERROR": {
		"Connection timeout after 30s retrying upstream service",
		"Failed to authenticate user: invalid token signature",
		"Disk space critical: 2%% remaining on volume /data",
		"Database connection pool exhausted (max: 50 connections)",
		"Unhandled exception in request handler: nil pointer dereference",
		"Elasticsearch cluster health: RED - %d unassigned shards",
		"Rate limit exceeded for API key, throttling requests",
		"TLS handshake failed: certificate expired for host %s",
		"Out of memory: killed process %s (pid %d)",
		"Index corruption detected in shard %d, initiating recovery",
	},
	"WARN": {
		"Memory usage at 85%%, consider scaling horizontally",
		"Slow query detected: %.1fs for index scan on %s",
		"Retry attempt %d/%d for failed message delivery",
		"Deprecated API endpoint called by client %s",
		"Replication lag detected: slave behind by %dms",
		"Heap usage at 75%%, GC pause time increasing",
		"Certificate for %s expires in %d days",
		"High number of 4xx errors: %d in last 5 minutes",
		"Disk I/O latency above threshold: %dms avg",
		"Connection pool usage at %d%% capacity",
	},
	"INFO": {
		"Service started successfully on port %d",
		"User %s logged in from IP %s",
		"Index pattern updated: %d new documents indexed",
		"Scheduled maintenance window starting",
		"Deployment v%s rolled out to production",
		"Health check passed: all services operational",
		"Cache warmed with %d entries from database",
		"Webhook sent to %s: status 200 OK",
		"Backup completed: %d GB in %d seconds",
		"New node joined the cluster: %s",
	},
	"DEBUG": {
		"Request processed in %dms: GET /api/search?q=%s",
		"Query execution plan: index scan on %s, filtered by %s",
		"Cache miss for key %s, fetching from origin",
		"Serializing response with %d fields",
		"Connection acquired from pool in %dms (pool size: %d)",
		"TLS session reused for connection to %s",
		"Garbage collection completed: freed %d MB in %dms",
		"Span [%s] started for request %s",
	},
	"TRACE": {
		"Entering function handleRequest with context",
		"Allocated %d bytes for request buffer",
		"Exiting function handleRequest after %dμs",
		"Sending %d bytes over wire to %s",
		"Decoding frame: type=%d length=%d flags=%d",
		"Lock acquired for mutex %s after %dμs wait",
		"File descriptor %d opened for reading: %s",
	},
}

func generateSampleData(count int) []LogEntry {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	entries := make([]LogEntry, 0, count)
	now := time.Now()
	apps := []string{"web", "api", "worker", "scheduler", "indexer"}
	envs := []string{"prod", "staging", "dev"}

	for i := range count {
		level := levels[rng.Intn(len(levels))]
		source := sources[rng.Intn(len(sources))]
		msgList := messages[level]
		msgTmpl := msgList[rng.Intn(len(msgList))]

		msg := fmt.Sprintf(msgTmpl,
			rng.Intn(90)+10,
			rng.Intn(5)+1,
			fmt.Sprintf("ip-10-0-%d-%d", rng.Intn(256), rng.Intn(256)),
			fmt.Sprintf("srv-%03d", rng.Intn(50)),
			fmt.Sprintf("%d.%d.%d", rng.Intn(10), rng.Intn(10), rng.Intn(5)),
			fmt.Sprintf("%.1f", float64(rng.Intn(50))/10+0.1),
			fmt.Sprintf("logs-%s-%s", apps[rng.Intn(len(apps))], envs[rng.Intn(len(envs))]),
		)

		ts := now.Add(-time.Duration(count-i) * time.Second).
			Add(-time.Duration(rng.Intn(1000)) * time.Millisecond)

		fields := map[string]any{
			"host":         fmt.Sprintf("ip-10-0-%d-%d", rng.Intn(256), rng.Intn(256)),
			"pid":          rng.Intn(50000) + 1000,
			"service.name": source,
			"environment":  envs[rng.Intn(len(envs))],
			"app":          apps[rng.Intn(len(apps))],
		}

		entries = append(entries, LogEntry{
			Timestamp: ts,
			Level:     level,
			Message:   msg,
			Source:    source,
			Fields:    fields,
		})
	}

	return entries
}

func loadFromFile(path string) ([]LogEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var entries []LogEntry
	if err := json.Unmarshal(data, &entries); err == nil {
		return entries, nil
	}

	var wrapped struct {
		Entries []LogEntry `json:"entries"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return wrapped.Entries, nil
}

type ESConfig struct {
	Address  string
	Index    string
	Username string
	Password string
	Size     int
}

func queryES(cfg ESConfig) ([]LogEntry, error) {
	esCfg := elasticsearch.Config{
		Addresses: []string{cfg.Address},
	}
	if cfg.Username != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("creating ES client: %w", err)
	}

	pingRes, err := es.Ping()
	if err != nil {
		return nil, fmt.Errorf("connecting to ES: %w", err)
	}
	defer pingRes.Body.Close()

	if pingRes.IsError() {
		return nil, fmt.Errorf("ES ping failed: %s", pingRes.String())
	}

	query := `{
		"size": %d,
		"sort": [{ "@timestamp": { "order": "desc" } }],
		"query": { "match_all": {} }
	}`

	res, err := es.Search(
		es.Search.WithIndex(cfg.Index),
		es.Search.WithBody(strings.NewReader(fmt.Sprintf(query, cfg.Size))),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		return nil, fmt.Errorf("ES search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES search error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding ES response: %w", err)
	}

	var entries []LogEntry
	for _, hit := range result.Hits.Hits {
		entry := mapToLogEntry(hit.Source)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	return entries, nil
}

func mapToLogEntry(source map[string]interface{}) *LogEntry {
	if source == nil {
		return nil
	}

	entry := &LogEntry{
		Fields: source,
	}

	// Timestamp: try common field names
	for _, key := range []string{"@timestamp", "timestamp", "time", "date"} {
		if v, ok := source[key]; ok {
			switch t := v.(type) {
			case string:
				if parsed, err := time.Parse(time.RFC3339, t); err == nil {
					entry.Timestamp = parsed
					break
				}
				if parsed, err := time.Parse("2006-01-02T15:04:05.000Z", t); err == nil {
					entry.Timestamp = parsed
					break
				}
			case float64:
				entry.Timestamp = time.UnixMilli(int64(t))
				break
			}
		}
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Level
	for _, key := range []string{"log.level", "level", "severity", "loglevel"} {
		if v, ok := source[key]; ok {
			entry.Level = fmt.Sprintf("%v", v)
			break
		}
	}
	if entry.Level == "" {
		entry.Level = "INFO"
	}
	entry.Level = strings.ToUpper(entry.Level)

	// Message
	for _, key := range []string{"message", "msg", "log.message", "event.message"} {
		if v, ok := source[key]; ok {
			entry.Message = fmt.Sprintf("%v", v)
			break
		}
	}
	if entry.Message == "" {
		entry.Message = "(no message)"
	}

	// Source
	for _, key := range []string{"host.name", "host", "source", "service.name", "service", "agent.name"} {
		if v, ok := source[key]; ok {
			entry.Source = fmt.Sprintf("%v", v)
			break
		}
	}
	if entry.Source == "" {
		entry.Source = "unknown"
	}

	return entry
}

func filterEntries(entries []LogEntry, query, levelFilter string) []LogEntry {
	query = strings.ToLower(strings.TrimSpace(query))
	levelFilter = strings.ToUpper(strings.TrimSpace(levelFilter))

	if query == "" && (levelFilter == "" || levelFilter == "ALL") {
		return entries
	}

	result := make([]LogEntry, 0, len(entries))
	for _, e := range entries {
		if levelFilter != "" && levelFilter != "ALL" && e.Level != levelFilter {
			continue
		}
		if query != "" {
			lower := strings.ToLower(e.Message + " " + e.Source + " " + e.Level)
			if !strings.Contains(lower, query) {
				continue
			}
		}
		result = append(result, e)
	}

	return result
}
