package util

import (
	"net"
	"strings"
)

// GeoInfo holds parsed geographic data from an IP address.
type GeoInfo struct {
	Country string
	City    string
}

// LookupIP attempts to determine geographic location from an IP address.
// Uses an offline approach with common private/loopback ranges and
// provides a basic framework that can be extended with MaxMind GeoLite2.
func LookupIP(ipStr string) GeoInfo {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return GeoInfo{}
	}

	// Private / loopback ranges
	if isPrivateOrLoopback(ip) {
		return GeoInfo{Country: "Local", City: "Private Network"}
	}

	// Cloud provider / hosting ranges (common VPS IPs)
	if isHostingProvider(ip) {
		return GeoInfo{Country: "Cloud", City: "Data Center"}
	}

	// China IP ranges (simplified detection)
	if isChinaIP(ip) {
		return GeoInfo{Country: "CN", City: ""}
	}

	return GeoInfo{Country: "Unknown", City: ""}
}

func isPrivateOrLoopback(ip net.IP) bool {
	// Loopback
	if ip.IsLoopback() {
		return true
	}
	// Link-local
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	// Private ranges
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}
	for _, cidr := range privateBlocks {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip) {
			return true
		}
	}
	return false
}

func isHostingProvider(ip net.IP) bool {
	// Common cloud/VPS IP ranges (simplified)
	hostingBlocks := []string{
		// AWS
		"3.0.0.0/8", "18.0.0.0/8", "35.80.0.0/12", "52.0.0.0/8", "54.0.0.0/8",
		// GCP
		"34.0.0.0/8", "35.192.0.0/12",
		// Azure
		"13.64.0.0/11", "20.0.0.0/8", "40.64.0.0/10",
		// Alibaba Cloud
		"47.88.0.0/15", "47.96.0.0/12", "120.24.0.0/14",
		// Tencent Cloud
		"43.128.0.0/10", "114.132.0.0/16",
		// DigitalOcean
		"64.23.0.0/16", "134.122.0.0/16",
	}
	for _, cidr := range hostingBlocks {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip) {
			return true
		}
	}
	return false
}

func isChinaIP(ip net.IP) bool {
	// IPv4 mapped as IPv6
	ip4 := ip.To4()
	if ip4 == nil {
		// For IPv6, check common Chinese IPv6 prefixes
		ipStr := ip.String()
		cnV6Prefixes := []string{
			"2408:", "2409:", "240e:", "240a:", "2001:250:", "2001:da8:",
		}
		for _, prefix := range cnV6Prefixes {
			if strings.HasPrefix(ipStr, prefix) {
				return true
			}
		}
		return false
	}

	// Simplified China IP detection for common ranges
	cnBlocks := []string{
		// Major China IP blocks
		"1.0.0.0/8", "14.0.0.0/8", "27.0.0.0/8", "36.0.0.0/8",
		"39.0.0.0/8", "42.0.0.0/8", "49.0.0.0/8", "58.0.0.0/8",
		"59.0.0.0/8", "60.0.0.0/8", "61.0.0.0/8",
		"101.0.0.0/8", "106.0.0.0/8", "110.0.0.0/8",
		"111.0.0.0/8", "112.0.0.0/8", "113.0.0.0/8",
		"114.0.0.0/8", "115.0.0.0/8", "116.0.0.0/8",
		"117.0.0.0/8", "118.0.0.0/8", "119.0.0.0/8",
		"120.0.0.0/8", "121.0.0.0/8", "122.0.0.0/8", "123.0.0.0/8",
		"124.0.0.0/8", "125.0.0.0/8",
		"171.0.0.0/8", "175.0.0.0/8",
		"180.0.0.0/8", "182.0.0.0/8", "183.0.0.0/8",
		"202.0.0.0/8", "203.0.0.0/8",
		"210.0.0.0/8", "211.0.0.0/8",
		"218.0.0.0/8", "219.0.0.0/8",
		"220.0.0.0/8", "221.0.0.0/8", "222.0.0.0/8", "223.0.0.0/8",
	}

	for _, cidr := range cnBlocks {
		_, block, _ := net.ParseCIDR(cidr)
		if block != nil && block.Contains(ip4) {
			return true
		}
	}
	return false
}
