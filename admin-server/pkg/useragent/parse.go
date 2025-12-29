package useragent

import (
	"strings"
)

// ParseUserAgent 解析 User-Agent 字符串，提取浏览器和操作系统信息
func ParseUserAgent(userAgent string) (browser, os string) {
	if userAgent == "" {
		return "未知", "未知"
	}

	ua := strings.ToLower(userAgent)

	// 解析浏览器
	browser = parseBrowser(ua)

	// 解析操作系统
	os = parseOS(ua)

	return browser, os
}

// parseBrowser 解析浏览器类型
func parseBrowser(ua string) string {
	switch {
	case strings.Contains(ua, "edg/"):
		return "Edge"
	case strings.Contains(ua, "chrome/") && !strings.Contains(ua, "edg/"):
		return "Chrome"
	case strings.Contains(ua, "firefox/"):
		return "Firefox"
	case strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome/"):
		return "Safari"
	case strings.Contains(ua, "opera/") || strings.Contains(ua, "opr/"):
		return "Opera"
	case strings.Contains(ua, "msie") || strings.Contains(ua, "trident/"):
		return "IE"
	default:
		return "未知"
	}
}

// parseOS 解析操作系统类型
func parseOS(ua string) string {
	switch {
	case strings.Contains(ua, "windows nt 10"):
		return "Windows 10"
	case strings.Contains(ua, "windows nt 6.3"):
		return "Windows 8.1"
	case strings.Contains(ua, "windows nt 6.2"):
		return "Windows 8"
	case strings.Contains(ua, "windows nt 6.1"):
		return "Windows 7"
	case strings.Contains(ua, "windows nt 6.0"):
		return "Windows Vista"
	case strings.Contains(ua, "windows nt 5.1"):
		return "Windows XP"
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os x"):
		return "macOS"
	case strings.Contains(ua, "iphone"):
		return "iOS"
	case strings.Contains(ua, "ipad"):
		return "iPadOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "linux"):
		return "Linux"
	case strings.Contains(ua, "unix"):
		return "Unix"
	default:
		return "未知"
	}
}
