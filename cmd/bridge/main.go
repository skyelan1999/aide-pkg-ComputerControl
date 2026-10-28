package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type request struct {
	Action      string   `json:"action"`
	AllowedApps []string `json:"allowedApps"`
	X           int      `json:"x"`
	Y           int      `json:"y"`
	Text        string   `json:"text"`
	Key         string   `json:"key"`
	Approved    bool     `json:"approved"`
}
type desktop interface {
	Execute(context.Context, request) (json.RawMessage, error)
}
type nativeDesktop struct{ helper string }

func (d nativeDesktop) Execute(ctx context.Context, req request) (json.RawMessage, error) {
	data, _ := json.Marshal(req)
	cmd := exec.CommandContext(ctx, d.helper, "--request")
	cmd.Stdin = bytes.NewReader(data)
	var out cappedBuffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, errors.New("Windows desktop helper failed or timed out; check .NET Framework 4.6.2+ and the interactive session")
	}
	if !json.Valid(out.Bytes()) {
		return nil, errors.New("Invalid desktop helper response")
	}
	return out.Bytes(), nil
}

type cappedBuffer struct{ bytes.Buffer }

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 20<<20 {
		return 0, errors.New("desktop response exceeds limit")
	}
	return b.Buffer.Write(p)
}

type bridge struct {
	token   string
	desktop desktop
	gate    chan struct{}
}

func allowedPeer(remote string) bool {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	for _, cidr := range []string{"172.16.0.0/12", "192.168.65.0/24"} {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
func reply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func (b *bridge) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fail := func(code int, msg string) { reply(w, code, map[string]any{"ok": false, "error": msg}) }
	if !allowedPeer(r.RemoteAddr) || r.Header.Get("Origin") != "" {
		fail(403, "Only local/Aide Docker clients are allowed")
		return
	}
	if len(b.token) < 32 || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+b.token)) != 1 {
		fail(401, "Unauthorized")
		return
	}
	if r.Method == "GET" && r.URL.Path == "/healthz" {
		reply(w, 200, map[string]any{"ok": true, "service": "aide-computer-bridge", "platform": runtime.GOOS, "version": "1.2.0", "capabilities": []string{"inspect", "window-snapshot", "click", "unicode-type", "key"}, "powershellRequired": false})
		return
	}
	actions := map[string]string{"/v1/inspect": "inspect", "/v1/snapshot": "snapshot", "/v1/click": "click", "/v1/type": "type", "/v1/key": "key"}
	action, ok := actions[r.URL.Path]
	if !ok || r.Method != "POST" {
		fail(404, "Unsupported operation")
		return
	}
	var req request
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		fail(400, "Invalid or oversized JSON request")
		return
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		fail(400, "Only one request is allowed")
		return
	}
	if req.Action != "" {
		fail(400, "Action is derived from the route")
		return
	}
	req.Action = action
	if len(req.AllowedApps) == 0 || len(req.AllowedApps) > 50 {
		fail(403, "Configure allowed application executable names first")
		return
	}
	for _, name := range req.AllowedApps {
		if strings.TrimSpace(name) == "" || len(name) > 120 {
			fail(400, "Invalid allowed application")
			return
		}
	}
	if action == "click" || action == "type" || action == "key" {
		if !req.Approved {
			fail(403, "Aide confirmation is required")
			return
		}
	}
	if len([]rune(req.Text)) > 2000 || len(req.Text) > 8000 {
		fail(400, "Text is too long")
		return
	}
	if action == "key" && !map[string]bool{"ENTER": true, "TAB": true, "ESCAPE": true, "BACKSPACE": true, "DELETE": true, "UP": true, "DOWN": true, "LEFT": true, "RIGHT": true, "SPACE": true}[req.Key] {
		fail(400, "Unsupported key")
		return
	}
	select {
	case b.gate <- struct{}{}:
		defer func() { <-b.gate }()
	default:
		fail(409, "A desktop operation is already running")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	result, err := b.desktop.Execute(ctx, req)
	if err != nil {
		fail(502, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(result)
}
func configureToken(dir string) (string, error) {
	if _, err := os.Stat(filepath.Join(dir, "compose.yaml")); err != nil {
		return "", errors.New("Select the Aide installation folder containing compose.yaml")
	}
	path := filepath.Join(dir, ".env")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	const key = "AIDE_COMPUTER_BRIDGE_TOKEN="
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimPrefix(line, "\ufeff"))
		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == strings.TrimSuffix(key, "=") {
			token := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			if len(token) >= 32 {
				return token, nil
			}
			return "", errors.New("Existing bridge token is invalid; repair it in Aide configuration")
		}
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	// The user's explicit folder selection pairs this bridge. Never print the token.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.WriteString("\r\n" + key + token + "\r\n")
	return token, err
}
func main() {
	dir := flag.String("aide-dir", "", "Aide installation folder (otherwise select in a native dialog)")
	flag.Parse()
	exe, _ := os.Executable()
	helper := filepath.Join(filepath.Dir(exe), "AideComputerDesktop.exe")
	if *dir == "" {
		out, err := exec.Command(helper, "--pick-folder").Output()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Folder selection cancelled or .NET Framework desktop helper unavailable")
			os.Exit(1)
		}
		var selection struct {
			Directory string `json:"directory"`
		}
		if json.Unmarshal(out, &selection) != nil || selection.Directory == "" {
			os.Exit(1)
		}
		*dir = selection.Directory
	}
	token, err := configureToken(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	handler := &bridge{token: token, desktop: nativeDesktop{helper}, gate: make(chan struct{}, 1)}
	server := &http.Server{Addr: "0.0.0.0:17778", Handler: handler, ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 8192}
	fmt.Println("Aide Windows Computer Bridge 1.2.0. Paired with the selected installation. Restart Aide once to load its token. Close this window to stop desktop access. No PowerShell is used.")
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
