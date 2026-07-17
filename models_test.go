package oneforall_test

import (
	"testing"

	oneforall "github.com/jiaohoping/oneforall_go"
)

func TestSubdomain_IPs(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want []string
	}{
		{"empty", "", nil},
		{"single", "1.2.3.4", []string{"1.2.3.4"}},
		{"multiple", "1.2.3.4,5.6.7.8", []string{"1.2.3.4", "5.6.7.8"}},
		{"spaces", "1.2.3.4, 5.6.7.8 ", []string{"1.2.3.4", "5.6.7.8"}},
		{"trailing comma", "1.2.3.4,", []string{"1.2.3.4"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := oneforall.Subdomain{IP: tt.ip}
			got := sub.IPs()
			if len(got) != len(tt.want) {
				t.Fatalf("IPs() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("IPs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestSubdomain_BoolAccessors(t *testing.T) {
	alive := oneforall.Subdomain{Alive: 1, Resolve: 1, Request: 1, CDN: 1, Find: 1}
	if !alive.IsAlive() {
		t.Error("IsAlive() should be true")
	}
	if !alive.IsResolved() {
		t.Error("IsResolved() should be true")
	}
	if !alive.IsRequested() {
		t.Error("IsRequested() should be true")
	}
	if !alive.IsCDN() {
		t.Error("IsCDN() should be true")
	}
	if !alive.IsNew() {
		t.Error("IsNew() should be true")
	}

	dead := oneforall.Subdomain{}
	if dead.IsAlive() {
		t.Error("IsAlive() should be false for zero value")
	}
	if dead.IsCDN() {
		t.Error("IsCDN() should be false for zero value")
	}
}

func TestResult_Filter(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", Alive: 1},
			{Subdomain: "b.example.com", Alive: 0},
			{Subdomain: "c.example.com", Alive: 1},
		},
	}
	got := r.Filter(func(s oneforall.Subdomain) bool { return s.IsAlive() })
	if len(got.Subdomains) != 2 {
		t.Fatalf("Filter(alive) returned %d subdomains, want 2", len(got.Subdomains))
	}
	for _, s := range got.Subdomains {
		if !s.IsAlive() {
			t.Errorf("non-alive subdomain %q leaked through filter", s.Subdomain)
		}
	}
}

func TestResult_Alive(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "alive.example.com", Alive: 1},
			{Subdomain: "dead.example.com", Alive: 0},
		},
	}
	got := r.Alive()
	if len(got.Subdomains) != 1 {
		t.Fatalf("Alive() returned %d subdomains, want 1", len(got.Subdomains))
	}
	if got.Subdomains[0].Subdomain != "alive.example.com" {
		t.Errorf("unexpected subdomain: %s", got.Subdomains[0].Subdomain)
	}
}

func TestResult_GroupByModule(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", Module: "dns"},
			{Subdomain: "b.example.com", Module: "brute"},
			{Subdomain: "c.example.com", Module: "dns"},
		},
	}
	grouped := r.GroupByModule()
	if len(grouped["dns"]) != 2 {
		t.Errorf("dns group has %d entries, want 2", len(grouped["dns"]))
	}
	if len(grouped["brute"]) != 1 {
		t.Errorf("brute group has %d entries, want 1", len(grouped["brute"]))
	}
}

func TestResult_Stats(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Alive: 1, CDN: 0, Resolve: 1, Find: 1, Module: "dns"},
			{Alive: 1, CDN: 1, Resolve: 1, Find: 0, Module: "dns"},
			{Alive: 0, CDN: 0, Resolve: 0, Find: 0, Module: "brute"},
		},
	}
	stats := r.Stats()
	if stats.Total != 3 {
		t.Errorf("Total = %d, want 3", stats.Total)
	}
	if stats.Alive != 2 {
		t.Errorf("Alive = %d, want 2", stats.Alive)
	}
	if stats.CDN != 1 {
		t.Errorf("CDN = %d, want 1", stats.CDN)
	}
	if stats.Resolved != 2 {
		t.Errorf("Resolved = %d, want 2", stats.Resolved)
	}
	if stats.New != 1 {
		t.Errorf("New = %d, want 1", stats.New)
	}
	if stats.ByModule["dns"] != 2 {
		t.Errorf("ByModule[dns] = %d, want 2", stats.ByModule["dns"])
	}
	if stats.ByModule["brute"] != 1 {
		t.Errorf("ByModule[brute] = %d, want 1", stats.ByModule["brute"])
	}
}

func TestResult_Filter_EmptyInput(t *testing.T) {
	r := oneforall.Result{}
	got := r.Filter(func(s oneforall.Subdomain) bool { return true })
	if len(got.Subdomains) != 0 {
		t.Errorf("Filter on empty result returned %d subdomains", len(got.Subdomains))
	}
}

func TestResult_Stats_Empty(t *testing.T) {
	r := oneforall.Result{}
	stats := r.Stats()
	if stats.Total != 0 {
		t.Errorf("Stats.Total = %d, want 0", stats.Total)
	}
	if len(stats.ByModule) != 0 {
		t.Errorf("Stats.ByModule should be empty map")
	}
}

// --- v0.3.0 new tests ---

func TestSubdomain_CNAMEs(t *testing.T) {
	tests := []struct {
		name  string
		cname string
		want  []string
	}{
		{"empty", "", nil},
		{"single", "cdn.example.com", []string{"cdn.example.com"}},
		{"multiple", "a.cdn.com,b.cdn.com", []string{"a.cdn.com", "b.cdn.com"}},
		{"spaces", "a.cdn.com, b.cdn.com ", []string{"a.cdn.com", "b.cdn.com"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := oneforall.Subdomain{CNAME: tt.cname}
			got := sub.CNAMEs()
			if len(got) != len(tt.want) {
				t.Fatalf("CNAMEs() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("CNAMEs()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestResult_Unique(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "www.example.com", Module: "dns"},
			{Subdomain: "mail.example.com", Module: "dns"},
			{Subdomain: "www.example.com", Module: "brute"}, // duplicate name
			{Subdomain: "api.example.com", Module: "cert"},
		},
	}
	got := r.Unique()
	if len(got.Subdomains) != 3 {
		t.Fatalf("Unique() returned %d subdomains, want 3", len(got.Subdomains))
	}
	// First occurrence should be kept (dns module for www.example.com)
	for _, s := range got.Subdomains {
		if s.Subdomain == "www.example.com" && s.Module != "dns" {
			t.Errorf("wrong occurrence kept for www.example.com: module=%s, want dns", s.Module)
		}
	}
}

func TestResult_Unique_Empty(t *testing.T) {
	r := oneforall.Result{}
	got := r.Unique()
	if len(got.Subdomains) != 0 {
		t.Errorf("Unique() on empty result returned %d subdomains", len(got.Subdomains))
	}
}

func TestResult_Unique_NoDuplicates(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.com"},
			{Subdomain: "b.com"},
			{Subdomain: "c.com"},
		},
	}
	got := r.Unique()
	if len(got.Subdomains) != 3 {
		t.Errorf("Unique() with no duplicates returned %d, want 3", len(got.Subdomains))
	}
}

func TestResult_GroupBySource(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", Source: "censys"},
			{Subdomain: "b.example.com", Source: "virustotal"},
			{Subdomain: "c.example.com", Source: "censys"},
		},
	}
	grouped := r.GroupBySource()
	if len(grouped["censys"]) != 2 {
		t.Errorf("censys group has %d entries, want 2", len(grouped["censys"]))
	}
	if len(grouped["virustotal"]) != 1 {
		t.Errorf("virustotal group has %d entries, want 1", len(grouped["virustotal"]))
	}
}

func TestResult_Stats_BySource(t *testing.T) {
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Module: "dns", Source: "axfr"},
			{Module: "dns", Source: "axfr"},
			{Module: "brute", Source: "wordlist"},
		},
	}
	stats := r.Stats()
	if stats.BySource["axfr"] != 2 {
		t.Errorf("BySource[axfr] = %d, want 2", stats.BySource["axfr"])
	}
	if stats.BySource["wordlist"] != 1 {
		t.Errorf("BySource[wordlist] = %d, want 1", stats.BySource["wordlist"])
	}
}

func TestResult_Stats_BySource_Empty(t *testing.T) {
	r := oneforall.Result{}
	stats := r.Stats()
	if stats.BySource == nil {
		t.Error("Stats.BySource should be an initialised map, not nil")
	}
}
