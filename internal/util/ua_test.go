package util

import "testing"

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		ua       string
		device   string
		browser  string
		os       string
	}{
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", "Desktop", "Chrome", "Windows 10"},
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15", "Desktop", "Safari", "macOS"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0", "Desktop", "Firefox", "Windows 10"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1", "Mobile", "Safari", "iOS"},
		{"Mozilla/5.0 (Linux; Android 14; Pixel 8) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36", "Mobile", "Chrome", "Android"},
		{"Googlebot/2.1 (+http://www.google.com/bot.html)", "Bot", "Other", "Unknown"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0", "Desktop", "Edge", "Windows 10"},
		{"", "Unknown", "Unknown", "Unknown"},
	}

	for _, tt := range tests {
		info := ParseUserAgent(tt.ua)
		if info.DeviceType != tt.device {
			t.Errorf("UA %q: device = %q, want %q", tt.ua, info.DeviceType, tt.device)
		}
		if info.Browser != tt.browser {
			t.Errorf("UA %q: browser = %q, want %q", tt.ua, info.Browser, tt.browser)
		}
		if info.OS != tt.os {
			t.Errorf("UA %q: os = %q, want %q", tt.ua, info.OS, tt.os)
		}
	}
}
