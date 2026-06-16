package mathchina

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Client fetches pages from mathchina.com with rate limiting and retries.
type Client struct {
	cfg  Config
	hc   *http.Client
	last time.Time
	mu   sync.Mutex
}

// NewClient builds a Client from cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		hc:  &http.Client{Timeout: cfg.Timeout},
	}
}

// GetPage fetches rawURL and returns the decoded body bytes (UTF-8).
func (c *Client) GetPage(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
		}
		body, ct, retry, err := c.do(ctx, rawURL)
		if err == nil {
			decoded, decErr := decodeBody(body, ct)
			if decErr != nil {
				return body, nil // return raw bytes if decode fails
			}
			return decoded, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("mathchina: GET %s: %w", rawURL, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, string, bool, error) {
	c.delay()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, "", true, err
	}
	defer func() { _ = resp.Body.Close() }()

	ct := resp.Header.Get("Content-Type")

	if resp.StatusCode >= 500 {
		return nil, ct, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ct, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ct, true, err
	}
	return body, ct, false, nil
}

// delay enforces the minimum delay between requests.
func (c *Client) delay() {
	if c.cfg.Delay <= 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if wait := c.cfg.Delay - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

// decodeBody converts GBK/GB2312 bodies to UTF-8.
func decodeBody(body []byte, contentType string) ([]byte, error) {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "gbk") || strings.Contains(ct, "gb2312") || strings.Contains(ct, "gb18030") {
		return simplifiedchinese.GBK.NewDecoder().Bytes(body)
	}
	// Check <meta charset> in first 2KB.
	head := strings.ToLower(string(body[:min(2048, len(body))]))
	if strings.Contains(head, `charset="gbk"`) || strings.Contains(head, `charset=gbk`) ||
		strings.Contains(head, `charset="gb2312"`) || strings.Contains(head, `charset=gb2312`) {
		return simplifiedchinese.GBK.NewDecoder().Bytes(body)
	}
	return body, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
