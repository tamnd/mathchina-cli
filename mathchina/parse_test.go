package mathchina

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

// --- ParseThreadList tests ---

func TestParseThreadList(t *testing.T) {
	body := loadFixture(t, "forum_list.html")
	threads, err := ParseThreadList(body, "7")
	if err != nil {
		t.Fatalf("ParseThreadList: %v", err)
	}
	// Should get 3 non-sticky threads.
	if len(threads) != 3 {
		t.Fatalf("want 3 threads, got %d", len(threads))
	}
	if threads[0].ID != "12345" {
		t.Errorf("thread[0].ID = %q, want 12345", threads[0].ID)
	}
	if !strings.Contains(threads[0].Title, "CMO") {
		t.Errorf("thread[0].Title = %q, want CMO in title", threads[0].Title)
	}
}

func TestParseThreadListReplyCounts(t *testing.T) {
	body := loadFixture(t, "forum_list.html")
	threads, err := ParseThreadList(body, "7")
	if err != nil {
		t.Fatalf("ParseThreadList: %v", err)
	}
	if threads[0].Replies != 12 {
		t.Errorf("thread[0].Replies = %d, want 12", threads[0].Replies)
	}
	if threads[0].Views != 1234 {
		t.Errorf("thread[0].Views = %d, want 1234", threads[0].Views)
	}
}

func TestParseThreadListDates(t *testing.T) {
	body := loadFixture(t, "forum_list.html")
	threads, err := ParseThreadList(body, "7")
	if err != nil {
		t.Fatalf("ParseThreadList: %v", err)
	}
	if threads[0].LastPostTime.IsZero() {
		t.Error("thread[0].LastPostTime should not be zero")
	}
}

func TestParseThreadListForumID(t *testing.T) {
	body := loadFixture(t, "forum_list.html")
	threads, _ := ParseThreadList(body, "15")
	for _, th := range threads {
		if th.ForumID != "15" {
			t.Errorf("ForumID = %q, want 15", th.ForumID)
		}
	}
}

func TestParseThreadListEmpty(t *testing.T) {
	body := []byte(`<html><body><table><tbody id="threadlisttableid"></tbody></table></body></html>`)
	threads, err := ParseThreadList(body, "7")
	if err != nil {
		t.Fatalf("ParseThreadList: %v", err)
	}
	if len(threads) != 0 {
		t.Errorf("expected empty slice, got %d", len(threads))
	}
}

// --- ParseThreadDetail tests ---

func TestParseThreadDetail(t *testing.T) {
	body := loadFixture(t, "thread_detail.html")
	p, err := ParseThreadDetail(body, "12345")
	if err != nil {
		t.Fatalf("ParseThreadDetail: %v", err)
	}
	if p.ID != "12345" {
		t.Errorf("ID = %q", p.ID)
	}
	if !strings.Contains(p.Title, "CMO") {
		t.Errorf("Title = %q, want CMO in title", p.Title)
	}
}

func TestParseThreadDetailContent(t *testing.T) {
	body := loadFixture(t, "thread_detail.html")
	p, err := ParseThreadDetail(body, "12345")
	if err != nil {
		t.Fatalf("ParseThreadDetail: %v", err)
	}
	if p.Content == "" {
		t.Error("Content should not be empty")
	}
	if !strings.Contains(p.Content, "三角形") {
		t.Errorf("Content should contain 三角形, got: %q", p.Content[:min(100, len(p.Content))])
	}
}

func TestParseThreadDetailPosts(t *testing.T) {
	body := loadFixture(t, "thread_detail.html")
	p, err := ParseThreadDetail(body, "12345")
	if err != nil {
		t.Fatalf("ParseThreadDetail: %v", err)
	}
	// Should have at least 1 reply post.
	if len(p.Posts) == 0 {
		t.Error("expected at least 1 reply post")
	}
}

// --- ParseSearchResults tests ---

func TestParseSearchResults(t *testing.T) {
	body := loadFixture(t, "search_results.html")
	threads, err := ParseSearchResults(body)
	if err != nil {
		t.Fatalf("ParseSearchResults: %v", err)
	}
	if len(threads) != 3 {
		t.Errorf("want 3 results, got %d", len(threads))
	}
	if threads[0].ID != "11111" {
		t.Errorf("thread[0].ID = %q, want 11111", threads[0].ID)
	}
}

func TestParseSearchResultsEmpty(t *testing.T) {
	body := []byte(`<html><body><p>没有搜索结果</p></body></html>`)
	threads, err := ParseSearchResults(body)
	if err != nil {
		t.Fatalf("ParseSearchResults: %v", err)
	}
	if len(threads) != 0 {
		t.Errorf("expected empty results, got %d", len(threads))
	}
}

// --- ParseForumIndex tests ---

func TestParseForumIndex(t *testing.T) {
	body := loadFixture(t, "forum_index.html")
	forums, err := ParseForumIndex(body)
	if err != nil {
		t.Fatalf("ParseForumIndex: %v", err)
	}
	if len(forums) < 1 {
		t.Fatalf("want at least 1 forum, got %d", len(forums))
	}
	found := false
	for _, f := range forums {
		if f.ID == "7" && strings.Contains(f.Name, "题库") {
			found = true
		}
	}
	if !found {
		t.Errorf("did not find forum fid=7 in %+v", forums)
	}
}

// --- parseDate tests ---

func TestParseDateFull(t *testing.T) {
	t0, err := parseDate("2024-1-5 08:30", time.Now())
	if err != nil {
		t.Fatalf("parseDate: %v", err)
	}
	if t0.Year() != 2024 || t0.Month() != 1 || t0.Day() != 5 {
		t.Errorf("parseDate = %v", t0)
	}
}

func TestParseDateDateOnly(t *testing.T) {
	t0, err := parseDate("2023-11-15", time.Now())
	if err != nil {
		t.Fatalf("parseDate: %v", err)
	}
	if t0.Year() != 2023 || t0.Month() != 11 || t0.Day() != 15 {
		t.Errorf("parseDate = %v", t0)
	}
}

func TestParseDateMinutesAgo(t *testing.T) {
	now := time.Now()
	t0, err := parseDate("5 分钟前", now)
	if err != nil {
		t.Fatalf("parseDate: %v", err)
	}
	diff := now.Sub(t0)
	if diff < 4*time.Minute || diff > 6*time.Minute {
		t.Errorf("parseDate 5分钟前: diff = %v, want ~5m", diff)
	}
}

// --- Client httptest tests ---

func TestClientGetPageSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		_, _ = w.Write(loadFixture(t, "forum_list.html"))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Delay = 0
	cfg.Retries = 0
	c := NewClient(cfg)

	body, err := c.GetPage(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if len(body) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestClientRetryOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("<html><body>ok</body></html>"))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.Delay = 0
	cfg.Retries = 3
	c := NewClient(cfg)

	body, err := c.GetPage(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("GetPage: %v", err)
	}
	if !strings.Contains(string(body), "ok") {
		t.Errorf("body = %q, want ok", body)
	}
	if hits != 3 {
		t.Errorf("server hits = %d, want 3", hits)
	}
}

func TestDecodeBodyGBK(t *testing.T) {
	// Simple GBK bytes for "你好" (ni hao)
	gbkBytes := []byte{0xC4, 0xE3, 0xBA, 0xC3} // GBK encoding of 你好
	decoded, err := decodeBody(gbkBytes, "text/html; charset=gbk")
	if err != nil {
		t.Fatalf("decodeBody: %v", err)
	}
	if string(decoded) != "你好" {
		t.Errorf("decoded = %q, want 你好", string(decoded))
	}
}

func TestDecodeBodyUTF8(t *testing.T) {
	body := []byte("<html>你好</html>")
	decoded, err := decodeBody(body, "text/html; charset=utf-8")
	if err != nil {
		t.Fatalf("decodeBody: %v", err)
	}
	if string(decoded) != string(body) {
		t.Errorf("UTF-8 body changed: %q", decoded)
	}
}

func TestDecodeBodyGBKMetaTag(t *testing.T) {
	gbkBytes := []byte(`<meta charset="gbk">` + "\xC4\xE3\xBA\xC3")
	decoded, err := decodeBody(gbkBytes, "text/html")
	if err != nil {
		t.Fatalf("decodeBody: %v", err)
	}
	if !strings.Contains(string(decoded), "你好") {
		t.Errorf("decoded = %q, want 你好", string(decoded))
	}
}

func TestFallbackForums(t *testing.T) {
	if len(FallbackForums) == 0 {
		t.Error("FallbackForums should not be empty")
	}
	for _, f := range FallbackForums {
		if f.ID == "" || f.Name == "" || f.URL == "" {
			t.Errorf("incomplete fallback forum: %+v", f)
		}
	}
}
