package manager

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLatestFallsBackToWebRedirectOnRateLimit(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")

	m := Manager{Client: testClient(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case githubAPI:
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
		case githubLatestWebURL:
			w.Header().Set("Location", "https://github.com/"+Repository+"/releases/tag/v2.1.7")
			w.WriteHeader(http.StatusFound)
		default:
			t.Fatalf("unexpected URL: %s", r.URL)
		}
	})}

	rel, err := m.Latest(t.Context())
	if err != nil {
		t.Fatalf("expected fallback to succeed, got err: %v", err)
	}
	if rel.Tag != "v2.1.7" {
		t.Fatalf("expected v2.1.7, got %q", rel.Tag)
	}
}

func TestLatestRejectsUnsafeRedirectTarget(t *testing.T) {
	cases := []string{
		"https://evil.example/AlejoPalavecino/codex-engineering-harness/releases/tag/v2.1.7",
		"https://github.com/" + Repository + "/releases/tag/../main",
		"http://github.com/" + Repository + "/releases/tag/v2.1.7",
	}
	for _, loc := range cases {
		m := Manager{Client: testClient(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.String() {
			case githubAPI:
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
			case githubLatestWebURL:
				w.Header().Set("Location", loc)
				w.WriteHeader(http.StatusFound)
			}
		})}
		_, err := m.Latest(t.Context())
		if err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
			t.Fatalf("expected rate-limit error after rejected redirect %q, got: %v", loc, err)
		}
	}
}

func TestFetchJSONAttachesBearerOnlyForGitHubAPI(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "secret-test-token")
	var apiAuth, webAuth string

	m := Manager{Client: testClient(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case githubAPI:
			apiAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusTooManyRequests)
		case githubLatestWebURL:
			webAuth = r.Header.Get("Authorization")
			w.Header().Set("Location", "https://github.com/"+Repository+"/releases/tag/v2.1.7")
			w.WriteHeader(http.StatusFound)
		}
	})}

	rel, err := m.Latest(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.Tag != "v2.1.7" {
		t.Fatalf("unexpected tag: %q", rel.Tag)
	}
	if apiAuth != "Bearer secret-test-token" {
		t.Fatalf("expected bearer token on api.github.com, got %q", apiAuth)
	}
	if webAuth != "" {
		t.Fatalf("must not leak bearer token to github.com web redirect, got %q", webAuth)
	}
}

func TestRateLimitErrorFormatting(t *testing.T) {
	reset := time.Now().Add(20 * time.Minute).Truncate(time.Minute)
	m := Manager{Client: testClient(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.String() {
		case githubAPI:
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
		case githubLatestWebURL:
			w.WriteHeader(http.StatusBadGateway)
		}
	})}

	_, err := m.Latest(t.Context())
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	wantTime := fmt.Sprintf("a las %s", reset.Local().Format("15:04"))
	if !strings.Contains(msg, "GITHUB_TOKEN") || !strings.Contains(msg, wantTime) {
		t.Fatalf("unexpected rate-limit error message: %q (want %q)", msg, wantTime)
	}
}
