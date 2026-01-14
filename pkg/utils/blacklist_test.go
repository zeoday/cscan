package utils

import (
	"testing"
)

func TestBlacklistMatcher_Domain(t *testing.T) {
	rules := []string{
		"*.gov",
		"*.edu",
		"*.mil",
		"example.com",
	}
	matcher := NewBlacklistMatcher(rules)

	tests := []struct {
		target   string
		expected bool
	}{
		{"test.gov", true},
		{"sub.test.gov", true},
		{"www.example.edu", true},
		{"army.mil", true},
		{"example.com", true},
		{"sub.example.com", false}, // 精确匹配，不匹配子域名
		{"test.com", false},
		{"gov.test.com", false},
	}

	for _, tt := range tests {
		result := matcher.IsBlacklisted(tt.target)
		if result != tt.expected {
			t.Errorf("IsBlacklisted(%s) = %v, expected %v", tt.target, result, tt.expected)
		}
	}
}

func TestBlacklistMatcher_IP(t *testing.T) {
	rules := []string{
		"192.168.1.1",
		"10.0.0.0/8",
		"172.16.0.0/12",
	}
	matcher := NewBlacklistMatcher(rules)

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"192.168.1.2", false},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"172.32.0.1", false},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		result := matcher.IsIPBlacklisted(tt.ip)
		if result != tt.expected {
			t.Errorf("IsIPBlacklisted(%s) = %v, expected %v", tt.ip, result, tt.expected)
		}
	}
}

func TestBlacklistMatcher_Keyword(t *testing.T) {
	rules := []string{
		"*cdn*",
		"*test*",
	}
	matcher := NewBlacklistMatcher(rules)

	tests := []struct {
		target   string
		expected bool
	}{
		{"cdn.example.com", true},
		{"example-cdn.com", true},
		{"mycdnserver.net", true},
		{"test.example.com", true},
		{"example.test.com", true},
		{"testing.com", true},
		{"example.com", false},
	}

	for _, tt := range tests {
		result := matcher.IsBlacklisted(tt.target)
		if result != tt.expected {
			t.Errorf("IsBlacklisted(%s) = %v, expected %v", tt.target, result, tt.expected)
		}
	}
}

func TestBlacklistMatcher_URL(t *testing.T) {
	rules := []string{
		"*.gov",
		"192.168.1.1",
	}
	matcher := NewBlacklistMatcher(rules)

	tests := []struct {
		target   string
		expected bool
	}{
		{"https://www.test.gov/path", true},
		{"http://192.168.1.1:8080/api", true},
		{"https://example.com", false},
	}

	for _, tt := range tests {
		result := matcher.IsBlacklisted(tt.target)
		if result != tt.expected {
			t.Errorf("IsBlacklisted(%s) = %v, expected %v", tt.target, result, tt.expected)
		}
	}
}

func TestBlacklistMatcher_FilterTargets(t *testing.T) {
	rules := []string{
		"*.gov",
		"10.0.0.0/8",
	}
	matcher := NewBlacklistMatcher(rules)

	targets := []string{
		"example.com",
		"test.gov",
		"10.0.0.1",
		"8.8.8.8",
		"www.example.gov",
	}

	filtered := matcher.FilterTargets(targets)
	expected := []string{"example.com", "8.8.8.8"}

	if len(filtered) != len(expected) {
		t.Errorf("FilterTargets returned %d items, expected %d", len(filtered), len(expected))
		return
	}

	for i, v := range filtered {
		if v != expected[i] {
			t.Errorf("FilterTargets[%d] = %s, expected %s", i, v, expected[i])
		}
	}
}

func TestBlacklistMatcher_Empty(t *testing.T) {
	matcher := NewBlacklistMatcher(nil)
	if !matcher.IsEmpty() {
		t.Error("Empty matcher should return IsEmpty() = true")
	}

	matcher2 := NewBlacklistMatcher([]string{"# comment", "", "  "})
	if !matcher2.IsEmpty() {
		t.Error("Matcher with only comments should return IsEmpty() = true")
	}
}

func TestBlacklistMatcher_Comments(t *testing.T) {
	rules := []string{
		"# This is a comment",
		"*.gov",
		"  # Another comment",
		"example.com",
	}
	matcher := NewBlacklistMatcher(rules)

	if matcher.RuleCount() != 2 {
		t.Errorf("RuleCount() = %d, expected 2", matcher.RuleCount())
	}
}


func TestNewExcludeHostsMatcher(t *testing.T) {
	tests := []struct {
		name         string
		excludeHosts string
		expectNil    bool
		ruleCount    int
	}{
		{"empty string", "", true, 0},
		{"single IP", "192.168.1.1", false, 1},
		{"multiple IPs", "192.168.1.1,10.0.0.1", false, 2},
		{"CIDR", "10.0.0.0/8", false, 1},
		{"mixed", "192.168.1.1,10.0.0.0/8,172.16.0.0/12", false, 3},
		{"with spaces", " 192.168.1.1 , 10.0.0.1 ", false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewExcludeHostsMatcher(tt.excludeHosts)
			if tt.expectNil {
				if matcher != nil {
					t.Errorf("Expected nil matcher for %q", tt.excludeHosts)
				}
				return
			}
			if matcher == nil {
				t.Errorf("Expected non-nil matcher for %q", tt.excludeHosts)
				return
			}
			if matcher.RuleCount() != tt.ruleCount {
				t.Errorf("RuleCount() = %d, expected %d", matcher.RuleCount(), tt.ruleCount)
			}
		})
	}
}

func TestNewExcludeHostsMatcher_Filtering(t *testing.T) {
	matcher := NewExcludeHostsMatcher("192.168.1.1,10.0.0.0/8")

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"192.168.1.2", false},
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"8.8.8.8", false},
	}

	for _, tt := range tests {
		result := matcher.IsIPBlacklisted(tt.ip)
		if result != tt.expected {
			t.Errorf("IsIPBlacklisted(%s) = %v, expected %v", tt.ip, result, tt.expected)
		}
	}
}

func TestMergeMatchers(t *testing.T) {
	matcher1 := NewBlacklistMatcher([]string{"192.168.1.1", "*.gov"})
	matcher2 := NewBlacklistMatcher([]string{"10.0.0.0/8", "*.edu"})

	merged := MergeMatchers(matcher1, matcher2)

	tests := []struct {
		target   string
		expected bool
	}{
		{"192.168.1.1", true},
		{"10.0.0.1", true},
		{"test.gov", true},
		{"test.edu", true},
		{"example.com", false},
	}

	for _, tt := range tests {
		result := merged.IsBlacklisted(tt.target)
		if result != tt.expected {
			t.Errorf("IsBlacklisted(%s) = %v, expected %v", tt.target, result, tt.expected)
		}
	}
}

func TestMergeMatchers_WithNil(t *testing.T) {
	matcher1 := NewBlacklistMatcher([]string{"192.168.1.1"})

	merged := MergeMatchers(nil, matcher1, nil)

	if !merged.IsIPBlacklisted("192.168.1.1") {
		t.Error("Merged matcher should contain rules from non-nil matcher")
	}
}

func TestFilterAssetsByIP(t *testing.T) {
	matcher := NewExcludeHostsMatcher("192.168.1.0/24,10.0.0.0/8")

	hosts := []string{"example.com", "test.com", "blocked.com", "safe.com"}
	ipv4Map := map[string][]string{
		"example.com": {"8.8.8.8"},
		"test.com":    {"192.168.1.100"}, // Should be filtered
		"blocked.com": {"10.0.0.1"},      // Should be filtered
		"safe.com":    {"1.1.1.1"},
	}

	filtered := matcher.FilterAssetsByIP(hosts, ipv4Map)

	expected := []string{"example.com", "safe.com"}
	if len(filtered) != len(expected) {
		t.Errorf("FilterAssetsByIP returned %d items, expected %d", len(filtered), len(expected))
		return
	}

	for i, v := range filtered {
		if v != expected[i] {
			t.Errorf("FilterAssetsByIP[%d] = %s, expected %s", i, v, expected[i])
		}
	}
}
