package middleware

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	// Check for forwarded headers (if behind proxy)
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	// Clean up IP (take first one if comma-separated)
	if strings.Contains(ip, ",") {
		ip = strings.Split(ip, ",")[0]
	}

	return strings.TrimSpace(ip)
}
