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

// --- v0.4.0: Diff tests ---

func TestResult_Diff_Added(t *testing.T) {
	prev := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com"},
		},
	}
	curr := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com"},
			{Subdomain: "b.example.com"}, // new
		},
	}
	diff := curr.Diff(prev)
	if len(diff.Added) != 1 {
		t.Fatalf("Added = %d, want 1", len(diff.Added))
	}
	if diff.Added[0].Subdomain != "b.example.com" {
		t.Errorf("Added[0] = %q, want b.example.com", diff.Added[0].Subdomain)
	}
	if len(diff.Removed) != 0 {
		t.Errorf("Removed = %d, want 0", len(diff.Removed))
	}
	if len(diff.Changed) != 0 {
		t.Errorf("Changed = %d, want 0", len(diff.Changed))
	}
}

func TestResult_Diff_Removed(t *testing.T) {
	prev := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com"},
			{Subdomain: "b.example.com"},
		},
	}
	curr := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com"},
		},
	}
	diff := curr.Diff(prev)
	if len(diff.Removed) != 1 {
		t.Fatalf("Removed = %d, want 1", len(diff.Removed))
	}
	if diff.Removed[0].Subdomain != "b.example.com" {
		t.Errorf("Removed[0] = %q, want b.example.com", diff.Removed[0].Subdomain)
	}
	if len(diff.Added) != 0 {
		t.Errorf("Added = %d, want 0", len(diff.Added))
	}
}

func TestResult_Diff_Changed_IP(t *testing.T) {
	prev := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", IP: "1.2.3.4"},
		},
	}
	curr := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", IP: "5.6.7.8"},
		},
	}
	diff := curr.Diff(prev)
	if len(diff.Changed) != 1 {
		t.Fatalf("Changed = %d, want 1", len(diff.Changed))
	}
	if diff.Changed[0].Before.IP != "1.2.3.4" {
		t.Errorf("Changed[0].Before.IP = %q, want 1.2.3.4", diff.Changed[0].Before.IP)
	}
	if diff.Changed[0].After.IP != "5.6.7.8" {
		t.Errorf("Changed[0].After.IP = %q, want 5.6.7.8", diff.Changed[0].After.IP)
	}
}

func TestResult_Diff_Changed_Alive(t *testing.T) {
	prev := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", Alive: 1},
		},
	}
	curr := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", Alive: 0},
		},
	}
	diff := curr.Diff(prev)
	if len(diff.Changed) != 1 {
		t.Fatalf("Changed = %d, want 1 (alive changed)", len(diff.Changed))
	}
}

func TestResult_Diff_NoChange(t *testing.T) {
	prev := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", IP: "1.2.3.4", Alive: 1, Status: 200},
		},
	}
	curr := oneforall.Result{
		Subdomains: []oneforall.Subdomain{
			{Subdomain: "a.example.com", IP: "1.2.3.4", Alive: 1, Status: 200},
		},
	}
	diff := curr.Diff(prev)
	if len(diff.Added)+len(diff.Removed)+len(diff.Changed) != 0 {
		t.Errorf("expected empty diff, got added=%d removed=%d changed=%d",
			len(diff.Added), len(diff.Removed), len(diff.Changed))
	}
}

func TestResult_Diff_EmptyInputs(t *testing.T) {
	diff := oneforall.Result{}.Diff(oneforall.Result{})
	if len(diff.Added)+len(diff.Removed)+len(diff.Changed) != 0 {
		t.Error("diff of two empty results should be empty")
	}
}

// --- v0.4.0: ScanMeta tests ---

func TestResult_Meta_ZeroValue(t *testing.T) {
	// Manually constructed Results have zero-value Meta — should not panic.
	r := oneforall.Result{
		Subdomains: []oneforall.Subdomain{{Subdomain: "example.com"}},
	}
	if !r.Meta.StartedAt.IsZero() {
		t.Error("manually constructed Result.Meta.StartedAt should be zero")
	}
}
