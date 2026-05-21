package router

import (
	"net"
	"net/http"
)

func WithTrustedSubnet(cidr string) func(http.Handler) http.Handler {
	if cidr == "" {
		return func(h http.Handler) http.Handler { return h }
	}

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "trusted subnet misconfigured", http.StatusForbidden)
			})
		}
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				http.Error(w, "X-Real-IP header is required", http.StatusForbidden)
				return
			}
			ip := net.ParseIP(realIP)
			if ip == nil {
				http.Error(w, "X-Real-IP header is invalid", http.StatusForbidden)
				return
			}
			if !ipNet.Contains(ip) {
				http.Error(w, "agent IP is not in the trusted subnet", http.StatusForbidden)
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
