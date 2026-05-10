package fetcher

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchManifestUsesGitHubApiAndHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/kkaddal-bc/maestro/releases/latest":
			if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
				t.Fatalf("latest Accept = %q", got)
			}
			if got := r.Header.Get("User-Agent"); got != userAgent {
				t.Fatalf("latest User-Agent = %q", got)
			}
			_, _ = io.WriteString(w, `{"tag_name":"v1.2.3","assets":[{"id":101,"name":"manifest.json"}]}`)
		case "/repos/kkaddal-bc/maestro/releases/assets/101":
			if got := r.Header.Get("Accept"); got != "application/octet-stream" {
				t.Fatalf("asset Accept = %q", got)
			}
			if got := r.Header.Get("User-Agent"); got != userAgent {
				t.Fatalf("asset User-Agent = %q", got)
			}
			_, _ = io.WriteString(w, `{"version":"v1.2.3","skills":[{"name":"maestro-snap","description":"Capture"}]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)
	got, err := client.FetchManifest()
	if err != nil {
		t.Fatalf("FetchManifest() error = %v", err)
	}
	if got.Version != "v1.2.3" {
		t.Fatalf("Version = %q", got.Version)
	}
	if len(got.Skills) != 1 || got.Skills[0].Name != "maestro-snap" {
		t.Fatalf("Skills = %#v", got.Skills)
	}
}

func TestFetchSkillsArchiveUsesTagEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/kkaddal-bc/maestro/releases/tags/v1.2.3":
			if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
				t.Fatalf("tag Accept = %q", got)
			}
			_, _ = io.WriteString(w, `{"tag_name":"v1.2.3","assets":[{"id":202,"name":"skills.tar.gz"}]}`)
		case "/repos/kkaddal-bc/maestro/releases/assets/202":
			if got := r.Header.Get("Accept"); got != "application/octet-stream" {
				t.Fatalf("download Accept = %q", got)
			}
			_, _ = io.WriteString(w, "archive-bytes")
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)
	rc, err := client.FetchSkillsArchive("v1.2.3")
	if err != nil {
		t.Fatalf("FetchSkillsArchive() error = %v", err)
	}
	defer rc.Close()

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(got) != "archive-bytes" {
		t.Fatalf("archive = %q", string(got))
	}
}

func TestFetcherReturnsHelpfulErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/kkaddal-bc/maestro/releases/latest":
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, "missing")
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewWithBaseURL(server.URL)
	if _, err := client.FetchManifest(); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("FetchManifest() error = %v, want 404", err)
	}
}
