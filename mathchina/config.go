package mathchina

import "time"

// Config holds the HTTP client configuration.
type Config struct {
	BaseURL   string
	UserAgent string
	Timeout   time.Duration
	Delay     time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults for the mathchina client.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "http://www.mathchina.com/bbs",
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		Timeout:   30 * time.Second,
		Delay:     500 * time.Millisecond,
		Retries:   3,
	}
}
