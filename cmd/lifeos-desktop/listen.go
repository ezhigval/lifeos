package main

import (
	"fmt"
	"net"
)

// desktopPort is stable on purpose. WKWebView keeps localStorage per origin,
// and a random port is a new origin on every launch.
const (
	desktopPort     = 47621
	desktopPortSpan = 8
)

func listenDesktop(addr string) (net.Listener, error) {
	if addr != "" {
		return net.Listen("tcp", addr)
	}
	var first error
	for port := desktopPort; port < desktopPort+desktopPortSpan; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, nil
		}
		if first == nil {
			first = err
		}
	}
	return nil, fmt.Errorf("desktop ports %d-%d are busy: %w", desktopPort, desktopPort+desktopPortSpan-1, first)
}
