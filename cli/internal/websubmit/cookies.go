package websubmit

import (
	"net/http"
	"net/url"
	"strings"
)

func cookiesToHTTP(cookies []Cookie) []*http.Cookie {
	out := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		out = append(out, &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HttpOnly: cookie.HTTPOnly,
		})
	}
	return out
}

func (c *Client) captureResponseCookies(requestURL *url.URL, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		tracked := Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   cookie.Domain,
			Path:     cookie.Path,
			Expires:  cookie.Expires,
			Secure:   cookie.Secure,
			HTTPOnly: cookie.HttpOnly,
		}
		if tracked.Domain == "" && requestURL != nil {
			tracked.Domain = requestURL.Hostname()
		}
		if tracked.Path == "" {
			tracked.Path = "/"
		}
		c.trackedCookies[c.cookieKey(tracked)] = tracked
	}
}

func (c *Client) exportTrackedCookies(host string) []Cookie {
	out := make([]Cookie, 0)
	for _, cookie := range c.trackedCookies {
		if cookie.Domain == "" || strings.Contains(host, trimCookieDomain(cookie.Domain)) || strings.Contains(cookie.Domain, host) {
			out = append(out, cookie)
		}
	}
	return out
}

func (c *Client) cookieKey(cookie Cookie) string {
	return cookie.Domain + "|" + cookie.Path + "|" + cookie.Name
}

func trimCookieDomain(domain string) string {
	return strings.TrimPrefix(domain, ".")
}
