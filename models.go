package oneforall

import "strings"

// Result holds all subdomains discovered by a scan.
type Result struct {
	Subdomains []Subdomain `json:"subdomains"`
}

// Subdomain represents a single subdomain record from the OneForAll SQLite
// result database.
type Subdomain struct {
	ID         int     `gorm:"column:id;primaryKey" json:"id"`
	Alive      int     `gorm:"column:alive"         json:"alive"`
	Request    int     `gorm:"column:request"       json:"request"`
	Resolve    int     `gorm:"column:resolve"       json:"resolve"`
	URL        string  `gorm:"column:url"           json:"url"`
	Subdomain  string  `gorm:"column:subdomain"     json:"subdomain"`
	Port       int     `gorm:"column:port"          json:"port"`
	Level      int     `gorm:"column:level"         json:"level"`
	CNAME      string  `gorm:"column:cname"         json:"cname"`
	IP         string  `gorm:"column:ip"            json:"ip"`
	CDN        int     `gorm:"column:cdn"           json:"cdn"`
	Status     int     `gorm:"column:status"        json:"status"`
	Reason     string  `gorm:"column:reason"        json:"reason"`
	Title      string  `gorm:"column:title"         json:"title"`
	Banner     string  `gorm:"column:banner"        json:"banner"`
	Header     string  `gorm:"column:header"        json:"header"`
	History    string  `gorm:"column:history"       json:"history"`
	Response   string  `gorm:"column:response"      json:"response"`
	IPTimes    string  `gorm:"column:ip_times"      json:"ip_times"`
	CNAMETimes string  `gorm:"column:cname_times"   json:"cname_times"`
	TTL        string  `gorm:"column:ttl"           json:"ttl"`
	CIDR       string  `gorm:"column:cidr"          json:"cidr"`
	ASN        string  `gorm:"column:asn"           json:"asn"`
	Org        string  `gorm:"column:org"           json:"org"`
	Addr       string  `gorm:"column:addr"          json:"addr"`
	ISP        string  `gorm:"column:isp"           json:"isp"`
	Resolver   string  `gorm:"column:resolver"      json:"resolver"`
	Module     string  `gorm:"column:module"        json:"module"`
	Source     string  `gorm:"column:source"        json:"source"`
	Elapse     float64 `gorm:"column:elapse"        json:"elapse"`
	Find       int     `gorm:"column:find"          json:"find"`
}

// IPs returns the individual IP addresses for this subdomain.
// OneForAll stores multiple addresses as a comma-separated string.
func (s Subdomain) IPs() []string {
	return splitCSV(s.IP)
}

// CNAMEs returns the individual CNAME values for this subdomain.
// OneForAll may store multiple CNAMEs as a comma-separated string.
func (s Subdomain) CNAMEs() []string {
	return splitCSV(s.CNAME)
}

// IsAlive reports whether the subdomain is considered alive (Alive == 1).
func (s Subdomain) IsAlive() bool { return s.Alive == 1 }

// IsResolved reports whether DNS resolution succeeded (Resolve == 1).
func (s Subdomain) IsResolved() bool { return s.Resolve == 1 }

// IsRequested reports whether an HTTP request was made to this subdomain
// (Request == 1).
func (s Subdomain) IsRequested() bool { return s.Request == 1 }

// IsCDN reports whether the subdomain is fronted by a CDN (CDN == 1).
func (s Subdomain) IsCDN() bool { return s.CDN == 1 }

// IsNew reports whether the subdomain was newly discovered in this run
// (Find == 1).
func (s Subdomain) IsNew() bool { return s.Find == 1 }

// splitCSV splits a comma-separated string into a trimmed, non-empty slice.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// --- Result helper methods ---

// Filter returns a new Result containing only the subdomains for which
// predicate returns true.
func (r Result) Filter(predicate func(Subdomain) bool) Result {
	out := Result{}
	for _, sub := range r.Subdomains {
		if predicate(sub) {
			out.Subdomains = append(out.Subdomains, sub)
		}
	}
	return out
}

// Alive returns a new Result containing only subdomains with Alive == 1.
func (r Result) Alive() Result {
	return r.Filter(Subdomain.IsAlive)
}

// Unique returns a new Result with duplicate subdomain names removed. When
// duplicates exist (e.g. from multi-target or multi-module scans), the first
// occurrence is kept.
func (r Result) Unique() Result {
	seen := make(map[string]struct{}, len(r.Subdomains))
	out := Result{}
	for _, sub := range r.Subdomains {
		if _, exists := seen[sub.Subdomain]; !exists {
			seen[sub.Subdomain] = struct{}{}
			out.Subdomains = append(out.Subdomains, sub)
		}
	}
	return out
}

// GroupByModule returns a map from module name to the subdomains discovered
// by that module.
func (r Result) GroupByModule() map[string][]Subdomain {
	m := make(map[string][]Subdomain)
	for _, sub := range r.Subdomains {
		m[sub.Module] = append(m[sub.Module], sub)
	}
	return m
}

// GroupBySource returns a map from source name to the subdomains discovered
// via that source.
func (r Result) GroupBySource() map[string][]Subdomain {
	m := make(map[string][]Subdomain)
	for _, sub := range r.Subdomains {
		m[sub.Source] = append(m[sub.Source], sub)
	}
	return m
}

// ResultStats summarises key counts from a scan result.
type ResultStats struct {
	Total    int
	Alive    int
	CDN      int
	Resolved int
	New      int
	ByModule map[string]int
	BySource map[string]int
}

// Stats computes aggregate statistics over the result set.
func (r Result) Stats() ResultStats {
	stats := ResultStats{
		Total:    len(r.Subdomains),
		ByModule: make(map[string]int),
		BySource: make(map[string]int),
	}
	for _, sub := range r.Subdomains {
		if sub.IsAlive() {
			stats.Alive++
		}
		if sub.IsCDN() {
			stats.CDN++
		}
		if sub.IsResolved() {
			stats.Resolved++
		}
		if sub.IsNew() {
			stats.New++
		}
		stats.ByModule[sub.Module]++
		stats.BySource[sub.Source]++
	}
	return stats
}

// applyFilter removes subdomains that do not satisfy predicate, in place.
func (r *Result) applyFilter(predicate func(Subdomain) bool) {
	filtered := r.Subdomains[:0]
	for _, sub := range r.Subdomains {
		if predicate(sub) {
			filtered = append(filtered, sub)
		}
	}
	r.Subdomains = filtered
}
