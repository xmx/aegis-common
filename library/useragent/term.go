package useragent

import "strings"

func IsSupportedANSI(ua string) bool {
	ua = strings.ToLower(ua)
	for _, s := range []string{
		"curl", "httpie", "lwp-request", "wget", "python-requests", "python-httpx", "openbsd ftp",
		"powershell", "fetch", "aiohttp", "http_get", "xh", "nushell", "zig",
	} {
		if strings.Contains(ua, s) {
			return true
		}
	}

	return false
}
