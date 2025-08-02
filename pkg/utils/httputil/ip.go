// pkg/httputil/ip.go
package httputil

import (
    "net/http"
    "strings"
)

// GetClientIP finds the real client IP address, considering proxy headers.
// It is no longer a method, but a public function.
func GetClientIP(r *http.Request) string {
    // Check X-Forwarded-For header first
    forwarded := r.Header.Get("X-Forwarded-For")
    if forwarded != "" {
        // Get first IP if multiple
        ips := strings.Split(forwarded, ",")
        return strings.TrimSpace(ips[0])
    }

    // Check X-Real-IP header
    realIP := r.Header.Get("X-Real-IP")
    if realIP != "" {
        return realIP
    }

    // Fall back to RemoteAddr - but strip the port
    remoteAddr := r.RemoteAddr

    // Handle IPv6 addresses like [::1]:65384
    if strings.HasPrefix(remoteAddr, "[") {
        if idx := strings.LastIndex(remoteAddr, "]:"); idx != -1 {
            // Correctly handle IPv6 format, e.g., [::1]
            return remoteAddr[1:idx]
        }
    }
    
    // Handle IPv4 addresses like 127.0.0.1:65384
    if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
        return remoteAddr[:idx]
    }
    
    // If no port found, return as is
    return remoteAddr
}