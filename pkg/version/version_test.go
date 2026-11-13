package version

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

// A failed update check must not stop lightsout. CheckForUpdate used to call
// os.Exit on error, so it runs in a child copy of this test binary.
func TestCheckForUpdateContinuesWhenCheckFails(t *testing.T) {
	if os.Getenv("LIGHTSOUT_VERSION_TEST_CHILD") == "1" {
		Version = "v0.0.0-test" // a release build, so the check runs
		CheckForUpdate()
		return
	}

	// Proxy that passes release.Check's connectivity probe but rejects the
	// GitHub API call, so Check returns an error without real network access.
	var mu sync.Mutex
	var connectHosts []string
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			mu.Lock()
			connectHosts = append(connectHosts, r.Host)
			mu.Unlock()
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer proxy.Close()

	cmd := exec.Command(os.Args[0], "-test.run=^TestCheckForUpdateContinuesWhenCheckFails$")
	cmd.Env = append(os.Environ(),
		"LIGHTSOUT_VERSION_TEST_CHILD=1",
		"HTTP_PROXY="+proxy.URL, "http_proxy="+proxy.URL,
		"HTTPS_PROXY="+proxy.URL, "https_proxy="+proxy.URL,
		"NO_PROXY=", "no_proxy=",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("CheckForUpdate() with failing GitHub API: %v\n%s", err, out)
	}

	// Guards against a vacuous pass: the GitHub call must have been attempted.
	mu.Lock()
	defer mu.Unlock()
	if len(connectHosts) != 1 || connectHosts[0] != "api.github.com:443" {
		t.Fatalf("proxy CONNECT hosts = %q, want [\"api.github.com:443\"]", connectHosts)
	}
}

// recordingTransport records every request and fails it, so no test traffic
// reaches the network.
type recordingTransport struct{ urls []string }

func (rt *recordingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	rt.urls = append(rt.urls, r.URL.String())
	return nil, errors.New("network disabled in test")
}

// hangingTransport never answers; it only gives up when the request's context
// is cancelled, like a network that stalls instead of failing.
type hangingTransport struct{}

func (hangingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	<-r.Context().Done()
	return nil, r.Context().Err()
}

// A stalled network must not hang lightsout before it quits anything: the
// update check gives up after updateCheckTimeout.
func TestCheckForUpdateGivesUpOnStalledNetwork(t *testing.T) {
	originalTransport, originalTimeout, originalClientTimeout := http.DefaultTransport, updateCheckTimeout, http.DefaultClient.Timeout
	http.DefaultTransport = hangingTransport{}
	updateCheckTimeout = 100 * time.Millisecond
	Version = "v0.0.0-test" // a release build, so the check runs
	t.Cleanup(func() {
		http.DefaultTransport = originalTransport
		updateCheckTimeout = originalTimeout
		http.DefaultClient.Timeout = originalClientTimeout
		Version = devVersion
	})

	done := make(chan struct{})
	go func() {
		CheckForUpdate()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("CheckForUpdate() still running after 3s on a stalled network")
	}
}

// A dev build is not a release, so comparing it with the latest release would
// always report an update. The check must be skipped without any request.
func TestCheckForUpdateSkipsDevBuild(t *testing.T) {
	rt := &recordingTransport{}
	original := http.DefaultTransport
	http.DefaultTransport = rt
	t.Cleanup(func() { http.DefaultTransport = original })

	Version = "dev"
	CheckForUpdate()

	if len(rt.urls) != 0 {
		t.Fatalf("dev build made requests %q, want none", rt.urls)
	}
}
