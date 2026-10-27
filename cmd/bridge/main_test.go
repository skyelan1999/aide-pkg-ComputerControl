package main

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeDesktop struct {
	calls int
	got   request
}

func (d *fakeDesktop) Execute(_ context.Context, r request) (json.RawMessage, error) {
	d.calls++
	d.got = r
	return json.RawMessage(`{"ok":true}`), nil
}
func TestBridgeBoundaries(t *testing.T) {
	cases := []struct {
		path, body, peer, auth, origin string
		want                           int
	}{
		{"/healthz", "", "127.0.0.1:4", "yes", "", 200},
		{"/healthz", "", "10.1.1.1:4", "yes", "", 403},
		{"/healthz", "", "127.0.0.1:4", "no", "", 401},
		{"/healthz", "", "127.0.0.1:4", "yes", "https://example.com", 403},
		{"/v1/inspect", `{"allowedApps":[]}`, "127.0.0.1:4", "yes", "", 403},
		{"/v1/click", `{"allowedApps":["notepad"],"x":1,"y":2}`, "127.0.0.1:4", "yes", "", 403},
		{"/v1/type", `{"allowedApps":["notepad"],"approved":true,"text":"测试"}`, "172.24.1.2:4", "yes", "", 200},
		{"/v1/key", `{"allowedApps":["notepad"],"approved":true,"key":"ENTER|TAB"}`, "127.0.0.1:4", "yes", "", 400},
		{"/v1/inspect", `{"allowedApps":["notepad"],"command":"cmd.exe"}`, "127.0.0.1:4", "yes", "", 400},
		{"/v1/inspect", `{"allowedApps":["notepad"]} {}`, "127.0.0.1:4", "yes", "", 400},
	}
	for _, tc := range cases {
		t.Run(tc.path+tc.body+tc.peer+tc.auth+tc.origin, func(t *testing.T) {
			d := &fakeDesktop{}
			b := &bridge{strings.Repeat("x", 32), d, make(chan struct{}, 1)}
			method := "POST"
			if tc.path == "/healthz" {
				method = "GET"
			}
			r := httptest.NewRequest(method, tc.path, strings.NewReader(tc.body))
			r.RemoteAddr = tc.peer
			if tc.auth == "yes" {
				r.Header.Set("Authorization", "Bearer "+b.token)
			}
			r.Header.Set("Origin", tc.origin)
			w := httptest.NewRecorder()
			b.ServeHTTP(w, r)
			if w.Code != tc.want {
				t.Fatalf("got %d: %s", w.Code, w.Body.String())
			}
			if tc.want != 200 && d.calls != 0 {
				t.Fatal("denied request reached desktop")
			}
		})
	}
}
func TestPairPreservesConfiguration(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "compose.yaml"), []byte("services: {}"), 0600)
	original := "OTHER_KEY=preserve-me\n"
	os.WriteFile(filepath.Join(d, ".env"), []byte(original), 0600)
	a, err := configureToken(d)
	if err != nil || len(a) != 64 {
		t.Fatal(err)
	}
	b, err := configureToken(d)
	if err != nil || a != b {
		t.Fatal("token changed", err)
	}
	data, _ := os.ReadFile(filepath.Join(d, ".env"))
	if !strings.HasPrefix(string(data), original) || strings.Count(string(data), "AIDE_COMPUTER_BRIDGE_TOKEN=") != 1 {
		t.Fatal("unrelated configuration changed")
	}
}
