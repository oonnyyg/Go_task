package websocket

import (
	"bufio"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// rawHandshakeServer starts a TCP server that replies to the WebSocket
// opening handshake with the given Upgrade and Connection header field
// values. It gives full control over the exact response bytes, which is
// needed to exercise multi-token header values that net/http would not
// produce on its own.
func rawHandshakeServer(t *testing.T, upgrade, connection string) (url string, shutdown func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				req, err := http.ReadRequest(bufio.NewReader(c))
				if err != nil {
					return
				}
				resp := "HTTP/1.1 101 Switching Protocols\r\n" +
					"Upgrade: " + upgrade + "\r\n" +
					"Connection: " + connection + "\r\n" +
					"Sec-WebSocket-Accept: " + computeAcceptKey(req.Header.Get("Sec-Websocket-Key")) + "\r\n" +
					"\r\n"
				c.Write([]byte(resp))
				// Hold the connection open long enough for the
				// client to finish reading the handshake response.
				time.Sleep(200 * time.Millisecond)
			}(c)
		}
	}()
	return "ws://" + ln.Addr().String(), func() { ln.Close() }
}

// TestDialConnectionTokenList verifies that the client accepts a Connection
// header listing multiple tokens (e.g. "upgrade, keep-alive") as long as it
// contains "upgrade".
func TestDialConnectionTokenList(t *testing.T) {
	url, shutdown := rawHandshakeServer(t, "websocket", "upgrade, keep-alive")
	defer shutdown()

	d := Dialer{HandshakeTimeout: 5 * time.Second}
	ws, _, err := d.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial with Connection 'upgrade, keep-alive' failed: %v", err)
	}
	ws.Close()
}

// TestDialUpgradeTokenList verifies that the client accepts an Upgrade header
// listing multiple tokens (e.g. "foo, websocket") as long as it contains
// "websocket".
func TestDialUpgradeTokenList(t *testing.T) {
	url, shutdown := rawHandshakeServer(t, "foo, websocket", "upgrade")
	defer shutdown()

	d := Dialer{HandshakeTimeout: 5 * time.Second}
	ws, _, err := d.Dial(url, nil)
	if err != nil {
		t.Fatalf("Dial with Upgrade 'foo, websocket' failed: %v", err)
	}
	ws.Close()
}

// TestDialRejectsMissingToken verifies that a response whose Upgrade and
// Connection headers omit the required tokens is still rejected.
func TestDialRejectsMissingToken(t *testing.T) {
	url, shutdown := rawHandshakeServer(t, "foo, bar", "keep-alive")
	defer shutdown()

	d := Dialer{HandshakeTimeout: 5 * time.Second}
	_, _, err := d.Dial(url, nil)
	if err == nil {
		t.Fatalf("Dial with missing tokens unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "bad handshake") {
		t.Fatalf("expected bad handshake error, got %v", err)
	}
}
