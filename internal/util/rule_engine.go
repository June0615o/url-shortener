package util

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/panhao/url-shortener/internal/model"
)

// ApplyRules evaluates routing rules against the request and returns the target URL.
// Returns empty string if no rule matches.
func ApplyRules(rulesJSON string, r *http.Request) string {
	var rules []model.Rule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		return ""
	}

	// Parse request context
	ua := ParseUserAgent(r.UserAgent())
	ip := getIPFromRequest(r)
	geo := LookupIP(ip)

	for _, rule := range rules {
		if matchRule(rule, r, ua, geo) {
			return rule.Action.Target
		}
	}

	return ""
}

func matchRule(rule model.Rule, r *http.Request, ua UserAgentInfo, geo GeoInfo) bool {
	for _, cond := range rule.Conditions {
		if !matchCondition(cond, r, ua, geo) {
			return false
		}
	}
	return true
}

func matchCondition(cond model.Condition, r *http.Request, ua UserAgentInfo, geo GeoInfo) bool {
	value := strings.ToLower(cond.Value)
	fieldValue := ""

	switch cond.Field {
	case "country":
		fieldValue = strings.ToLower(geo.Country)
	case "city":
		fieldValue = strings.ToLower(geo.City)
	case "os":
		fieldValue = strings.ToLower(ua.OS)
	case "browser":
		fieldValue = strings.ToLower(ua.Browser)
	case "device":
		fieldValue = strings.ToLower(ua.DeviceType)
	case "language":
		fieldValue = strings.ToLower(r.Header.Get("Accept-Language"))
	case "referer":
		ref := r.Referer()
		if ref != "" {
			fieldValue = strings.ToLower(ref)
		}
	default:
		return false
	}

	switch cond.Op {
	case "eq":
		return fieldValue == value
	case "neq":
		return fieldValue != value
	case "contains":
		return strings.Contains(fieldValue, value)
	case "prefix":
		return strings.HasPrefix(fieldValue, value)
	case "regex":
		// Simple regex not implemented; always false for safety
		return false
	default:
		return false
	}
}

func getIPFromRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		return host[:idx]
	}
	return host
}

// BlacklistedDomains is a list of known malicious/spam domains.
var BlacklistedDomains = []string{
	"bit.ly", "tinyurl.com", "ow.ly", // competing shorteners
	"localhost", "127.0.0.1",
	"0.0.0.0",
}

// IsDomainBlacklisted checks if a URL's host is in the blacklist.
func IsDomainBlacklisted(rawURL string) bool {
	for _, domain := range BlacklistedDomains {
		if strings.Contains(strings.ToLower(rawURL), domain) {
			return true
		}
	}
	return false
}
