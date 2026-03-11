package websubmit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) fetchTimesheetHome(ctx context.Context) (page, error) {
	base := c.baseURLOrFallback("https://harvestapp.com")
	return c.getPage(ctx, strings.TrimRight(base, "/")+"/time")
}

func (c *Client) getPage(ctx context.Context, rawURL string) (page, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return page{}, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("User-Agent", c.UserAgent)

	response, err := c.HTTPClient.Do(request)
	if err != nil {
		return page{}, fmt.Errorf("request failed: %w", err)
	}
	defer response.Body.Close()
	c.captureResponseCookies(response.Request.URL, response.Cookies())

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return page{}, fmt.Errorf("read response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return page{}, fmt.Errorf("%s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	return parsePage(body, response.Request.URL)
}

func (c *Client) postForm(ctx context.Context, rawURL string, values url.Values, referer string) (page, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, strings.NewReader(values.Encode()))
	if err != nil {
		return page{}, fmt.Errorf("build request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Referer", referer)
	request.Header.Set("User-Agent", c.UserAgent)

	response, err := c.doRequestWithoutRedirects(request)
	if err != nil {
		return page{}, err
	}
	defer response.Body.Close()

	redirects := 0
	for response.StatusCode >= 300 && response.StatusCode < 400 {
		c.captureResponseCookies(response.Request.URL, response.Cookies())
		location := response.Header.Get("Location")
		if location == "" {
			break
		}
		if redirects >= 10 {
			return page{}, errors.New("too many redirects while submitting Harvest form")
		}
		redirects++

		nextURL, err := response.Request.URL.Parse(location)
		if err != nil {
			return page{}, fmt.Errorf("parse redirect url: %w", err)
		}
		method := http.MethodGet
		var body io.Reader
		if response.StatusCode == http.StatusTemporaryRedirect || response.StatusCode == http.StatusPermanentRedirect {
			method = request.Method
			body = strings.NewReader(values.Encode())
		}
		_ = response.Body.Close()

		request, err = http.NewRequestWithContext(ctx, method, nextURL.String(), body)
		if err != nil {
			return page{}, fmt.Errorf("build redirect request: %w", err)
		}
		request.Header.Set("User-Agent", c.UserAgent)
		if method == http.MethodPost {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Referer", referer)
		}

		response, err = c.doRequestWithoutRedirects(request)
		if err != nil {
			return page{}, err
		}
	}

	c.captureResponseCookies(response.Request.URL, response.Cookies())
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return page{}, fmt.Errorf("read response: %w", err)
	}
	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return page{}, fmt.Errorf("%s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	return parsePage(body, response.Request.URL)
}

func (c *Client) doRequestWithoutRedirects(request *http.Request) (*http.Response, error) {
	client := *c.HTTPClient
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return response, nil
}
