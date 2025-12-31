package githubnotifications

import (
	"strings"
	"testing"
	"time"

	"github.com/nbr23/rss-banquet/parser"
)

func TestConvertAPIURLToWebURL(t *testing.T) {
	testCases := []struct {
		apiURL  string
		webURL  string
	}{
		{
			"https://api.github.com/repos/owner/repo/issues/123",
			"https://github.com/owner/repo/issues/123",
		},
		{
			"https://api.github.com/repos/owner/repo/pulls/456",
			"https://github.com/owner/repo/pull/456",
		},
		{
			"https://api.github.com/repos/owner/repo/commits/abc123",
			"https://github.com/owner/repo/commit/abc123",
		},
		{
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.apiURL, func(t *testing.T) {
			result := convertAPIURLToWebURL(tc.apiURL)
			if result != tc.webURL {
				t.Errorf("got %q, wanted %q", result, tc.webURL)
			}
		})
	}
}

func TestFormatReason(t *testing.T) {
	testCases := []struct {
		reason   string
		expected string
	}{
		{"review_requested", "review requested"},
		{"state_change", "state change"},
		{"mention", "mention"},
		{"ci_activity", "ci activity"},
	}

	for _, tc := range testCases {
		t.Run(tc.reason, func(t *testing.T) {
			result := formatReason(tc.reason)
			if result != tc.expected {
				t.Errorf("got %q, wanted %q", result, tc.expected)
			}
		})
	}
}

func TestBuildItemTitle(t *testing.T) {
	n := &notification{
		Repository: notificationRepository{
			FullName: "owner/repo",
		},
		Subject: notificationSubject{
			Title: "Fix bug in parser",
		},
	}

	expected := "[owner/repo] Fix bug in parser"
	result := buildItemTitle(n)
	if result != expected {
		t.Errorf("got %q, wanted %q", result, expected)
	}
}

func TestFilterNotifications(t *testing.T) {
	notifications := []notification{
		{
			ID: "1",
			Repository: notificationRepository{
				Owner: notificationOwner{Login: "org1"},
			},
			Reason: "review_requested",
		},
		{
			ID: "2",
			Repository: notificationRepository{
				Owner: notificationOwner{Login: "org2"},
			},
			Reason: "mention",
		},
		{
			ID: "3",
			Repository: notificationRepository{
				Owner: notificationOwner{Login: "org1"},
			},
			Reason: "mention",
		},
	}

	t.Run("filter by org", func(t *testing.T) {
		options := &parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "org", Type: "string", Value: "org1"},
				&parser.Option{Flag: "reason", Type: "stringSlice", Value: ""},
			},
		}
		result := filterNotifications(notifications, options)
		if len(result) != 2 {
			t.Errorf("got %d items, wanted 2", len(result))
		}
	})

	t.Run("filter by reason", func(t *testing.T) {
		options := &parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "org", Type: "string", Value: ""},
				&parser.Option{Flag: "reason", Type: "stringSlice", Value: "mention"},
			},
		}
		result := filterNotifications(notifications, options)
		if len(result) != 2 {
			t.Errorf("got %d items, wanted 2", len(result))
		}
	})

	t.Run("filter by org and reason", func(t *testing.T) {
		options := &parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "org", Type: "string", Value: "org1"},
				&parser.Option{Flag: "reason", Type: "stringSlice", Value: "mention"},
			},
		}
		result := filterNotifications(notifications, options)
		if len(result) != 1 {
			t.Errorf("got %d items, wanted 1", len(result))
		}
	})

	t.Run("no filter", func(t *testing.T) {
		options := &parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "org", Type: "string", Value: ""},
				&parser.Option{Flag: "reason", Type: "stringSlice", Value: ""},
			},
		}
		result := filterNotifications(notifications, options)
		if len(result) != 3 {
			t.Errorf("got %d items, wanted 3", len(result))
		}
	})
}

func TestParseAge(t *testing.T) {
	testCases := []struct {
		age      string
		valid    bool
		checkFn  func(time.Time) bool
	}{
		{"24h", true, func(t time.Time) bool { return time.Since(t) >= 23*time.Hour && time.Since(t) <= 25*time.Hour }},
		{"7d", true, func(t time.Time) bool { return time.Since(t) >= 6*24*time.Hour && time.Since(t) <= 8*24*time.Hour }},
		{"2w", true, func(t time.Time) bool { return time.Since(t) >= 13*24*time.Hour && time.Since(t) <= 15*24*time.Hour }},
		{"", true, func(t time.Time) bool { return t.IsZero() }},
		{"invalid", false, nil},
		{"5x", false, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.age, func(t *testing.T) {
			result, err := parseAge(tc.age)
			if tc.valid {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tc.checkFn != nil && !tc.checkFn(result) {
					t.Errorf("time check failed for %q, got %v", tc.age, result)
				}
			} else {
				if err == nil {
					t.Errorf("expected error for %q", tc.age)
				}
			}
		})
	}
}

func TestBuildAPIURL(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		allFalse := false
		options := &parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "all", Type: "bool", Value: &allFalse},
				&parser.Option{Flag: "participating", Type: "bool", Value: &allFalse},
				&parser.Option{Flag: "age", Type: "string", Value: ""},
				&parser.Option{Flag: "before", Type: "string", Value: ""},
			},
		}
		result := buildAPIURL(options)
		expected := "https://api.github.com/notifications?per_page=50"
		if result != expected {
			t.Errorf("got %q, wanted %q", result, expected)
		}
	})

	t.Run("with all=true", func(t *testing.T) {
		allTrue := true
		allFalse := false
		options := &parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "all", Type: "bool", Value: &allTrue},
				&parser.Option{Flag: "participating", Type: "bool", Value: &allFalse},
				&parser.Option{Flag: "age", Type: "string", Value: ""},
				&parser.Option{Flag: "before", Type: "string", Value: ""},
			},
		}
		result := buildAPIURL(options)
		if result != "https://api.github.com/notifications?all=true&per_page=50" {
			t.Errorf("got %q", result)
		}
	})

	t.Run("with age", func(t *testing.T) {
		allFalse := false
		options := &parser.Options{
			OptionsList: parser.OptionsList{
				&parser.Option{Flag: "all", Type: "bool", Value: &allFalse},
				&parser.Option{Flag: "participating", Type: "bool", Value: &allFalse},
				&parser.Option{Flag: "age", Type: "string", Value: "24h"},
				&parser.Option{Flag: "before", Type: "string", Value: ""},
			},
		}
		result := buildAPIURL(options)
		// Should contain "since=" parameter
		if !strings.Contains(result, "since=") {
			t.Errorf("expected URL to contain 'since=' param, got %q", result)
		}
	})
}

func TestGetOptions(t *testing.T) {
	p := GitHubNotifications{}
	options := p.GetOptions()

	expectedFlags := []string{"all", "participating", "age", "before", "org", "reason", "title", "description"}
	for _, flag := range expectedFlags {
		found := false
		for _, opt := range options.OptionsList {
			if opt.Flag == flag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing expected option: %s", flag)
		}
	}
}

func TestString(t *testing.T) {
	p := GitHubNotifications{}
	if p.String() != "github-notifications" {
		t.Errorf("got %q, wanted %q", p.String(), "github-notifications")
	}
}
