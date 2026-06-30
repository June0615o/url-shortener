package util

import (
	"strings"
)

// UserAgentInfo holds parsed User-Agent data.
type UserAgentInfo struct {
	DeviceType string // "Desktop", "Mobile", "Tablet", "Bot", "Unknown"
	Browser    string // Chrome, Firefox, Safari, Edge, etc.
	OS         string // Windows, macOS, Linux, iOS, Android, etc.
}

// ParseUserAgent extracts device type, browser, and OS from a User-Agent string.
func ParseUserAgent(ua string) UserAgentInfo {
	info := UserAgentInfo{
		DeviceType: "Unknown",
		Browser:    "Unknown",
		OS:         "Unknown",
	}

	if ua == "" {
		return info
	}

	uaLower := strings.ToLower(ua)

	// Device type detection
	info.DeviceType = detectDeviceType(uaLower)

	// Browser detection
	info.Browser = detectBrowser(uaLower)

	// OS detection
	info.OS = detectOS(uaLower)

	return info
}

func detectDeviceType(ua string) string {
	// Bots first
	bots := []string{"bot", "crawler", "spider", "scraper", "slurp", "curl", "wget", "python-requests", "go-http-client", "httpclient"}
	for _, bot := range bots {
		if strings.Contains(ua, bot) {
			return "Bot"
		}
	}

	// Tablet
	tablets := []string{"ipad", "tablet", "kindle", "playbook", "silk"}
	for _, t := range tablets {
		if strings.Contains(ua, t) {
			return "Tablet"
		}
	}

	// Mobile
	mobiles := []string{"mobile", "iphone", "ipod", "android", "blackberry", "windows phone", "opera mini", "iemobile"}
	for _, m := range mobiles {
		if strings.Contains(ua, m) {
			return "Mobile"
		}
	}

	// Check for "Android" without "Mobile" — could be tablet
	if strings.Contains(ua, "android") {
		return "Mobile"
	}

	return "Desktop"
}

func detectBrowser(ua string) string {
	switch {
	case strings.Contains(ua, "edg/") || strings.Contains(ua, "edge/") || strings.Contains(ua, "edgios"):
		return "Edge"
	case strings.Contains(ua, "opr/") || strings.Contains(ua, "opera"):
		return "Opera"
	case strings.Contains(ua, "chrome/") && !strings.Contains(ua, "edg/") && !strings.Contains(ua, "opr/"):
		return "Chrome"
	case strings.Contains(ua, "safari/") && strings.Contains(ua, "applewebkit") && !strings.Contains(ua, "chrome"):
		return "Safari"
	case strings.Contains(ua, "firefox/"):
		return "Firefox"
	case strings.Contains(ua, "msie") || strings.Contains(ua, "trident/"):
		return "IE"
	case strings.Contains(ua, "wechat") || strings.Contains(ua, "micromessenger"):
		return "WeChat"
	case strings.Contains(ua, "qq/") || strings.Contains(ua, "mqqbrowser"):
		return "QQ Browser"
	case strings.Contains(ua, "ucbrowser") || strings.Contains(ua, "ucweb"):
		return "UC Browser"
	case strings.Contains(ua, "baidu") || strings.Contains(ua, "bidubrowser"):
		return "Baidu Browser"
	case strings.Contains(ua, "curl"):
		return "curl"
	case strings.Contains(ua, "wget"):
		return "wget"
	case strings.Contains(ua, "python"):
		return "Python"
	default:
		return "Other"
	}
}

func detectOS(ua string) string {
	switch {
	case strings.Contains(ua, "windows nt 11"):
		return "Windows 11"
	case strings.Contains(ua, "windows nt 10"):
		return "Windows 10"
	case strings.Contains(ua, "windows nt 6.3"):
		return "Windows 8.1"
	case strings.Contains(ua, "windows nt 6.2"):
		return "Windows 8"
	case strings.Contains(ua, "windows nt 6.1"):
		return "Windows 7"
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os x") || strings.Contains(ua, "macos"):
		return "macOS"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod"):
		return "iOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "linux") && !strings.Contains(ua, "android"):
		return "Linux"
	case strings.Contains(ua, "cros"):
		return "ChromeOS"
	case strings.Contains(ua, "freebsd"):
		return "FreeBSD"
	default:
		return "Unknown"
	}
}
