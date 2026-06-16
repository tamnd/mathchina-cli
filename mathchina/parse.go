package mathchina

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// baseURL is the Discuz forum root.
const baseURL = "http://www.mathchina.com/bbs"

// ParseThreadList parses a Discuz! forum listing page and returns threads.
func ParseThreadList(body []byte, forumID string) ([]Thread, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var threads []Thread
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tbody" {
			if attr(n, "id") == "threadlisttableid" {
				parseThreadListTable(n, forumID, &threads)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return threads, nil
}

func parseThreadListTable(tbody *html.Node, forumID string, out *[]Thread) {
	for row := tbody.FirstChild; row != nil; row = row.NextSibling {
		if row.Type != html.ElementNode || row.Data != "tr" {
			continue
		}
		// Skip header/sticky rows with class "bgm" (pinned threads)
		cl := attr(row, "class")
		if strings.Contains(cl, "bgm") {
			continue
		}
		t := parseThreadRow(row, forumID)
		if t.ID != "" {
			*out = append(*out, t)
		}
	}
}

func parseThreadRow(row *html.Node, forumID string) Thread {
	var t Thread
	t.ForumID = forumID

	// Collect all td elements.
	var tds []*html.Node
	for c := row.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "td" {
			tds = append(tds, c)
		}
	}

	for _, td := range tds {
		cl := attr(td, "class")
		switch {
		case strings.Contains(cl, "new") || strings.Contains(cl, "tit"):
			// Title column — find the thread link.
			findLinks(td, func(href, text string) bool {
				if idx := strings.Index(href, "tid="); idx >= 0 {
					part := href[idx+4:]
					if amp := strings.Index(part, "&"); amp >= 0 {
						part = part[:amp]
					}
					t.ID = part
					t.Title = strings.TrimSpace(text)
					t.URL = baseURL + "/forum.php?mod=viewthread&tid=" + t.ID
					return true
				}
				return false
			})
		case cl == "by":
			// Author or last-post column.
			if t.Author == "" {
				// First "by" column is the author.
				findLinks(td, func(href, text string) bool {
					if strings.Contains(href, "uid=") || strings.Contains(href, "space.php") {
						t.Author = strings.TrimSpace(text)
						return true
					}
					return false
				})
				// Also try <cite> text.
				if t.Author == "" {
					t.Author = strings.TrimSpace(textContent(findElement(td, "cite")))
				}
			} else {
				// Second "by" column is last post time.
				if em := findElement(td, "em"); em != nil {
					ts := strings.TrimSpace(textContent(em))
					if parsed, err := parseDate(ts, time.Now()); err == nil {
						t.LastPostTime = parsed
					}
				}
			}
		case cl == "num":
			// Reply/view counts.
			nums := collectText(td)
			if len(nums) >= 2 {
				t.Replies, _ = strconv.Atoi(cleanNum(nums[0]))
				t.Views, _ = strconv.Atoi(cleanNum(nums[1]))
			}
		}
	}
	return t
}

// ParseThreadDetail parses a Discuz! thread detail page.
func ParseThreadDetail(body []byte, threadID string) (*Problem, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	p := &Problem{
		ID:  threadID,
		URL: baseURL + "/forum.php?mod=viewthread&tid=" + threadID,
	}

	// Extract title from <h1 id="thread_subject"> or <span id="thread_subject">.
	if n := findByID(doc, "thread_subject"); n != nil {
		p.Title = strings.TrimSpace(textContent(n))
	}

	// Extract posts.
	var posts []Post
	var walkPosts func(*html.Node)
	walkPosts = func(n *html.Node) {
		if n.Type == html.ElementNode {
			id := attr(n, "id")
			if strings.HasPrefix(id, "postmessage_") {
				postID := strings.TrimPrefix(id, "postmessage_")
				content := cleanHTML(n)
				author := ""
				postTime := time.Time{}

				// Walk up to find the parent post container.
				parent := n.Parent
				for i := 0; i < 5 && parent != nil; i++ {
					// Look for author in <div class="authi"> or <td class="pls">.
					authorNode := findByClass(parent, "authi")
					if authorNode == nil {
						authorNode = findByClass(parent, "pls")
					}
					if authorNode != nil && author == "" {
						if a := findElement(authorNode, "a"); a != nil {
							author = strings.TrimSpace(textContent(a))
						}
					}
					// Look for time in <em class="authorinfo"> or timestamp.
					timeNode := findByClass(parent, "authorinfo")
					if timeNode == nil {
						timeNode = findByClass(parent, "pti")
					}
					if timeNode != nil && postTime.IsZero() {
						ts := strings.TrimSpace(textContent(timeNode))
						if t, err := parseDate(ts, time.Now()); err == nil {
							postTime = t
						}
					}
					parent = parent.Parent
				}

				posts = append(posts, Post{
					PostID:   postID,
					Author:   author,
					Content:  content,
					PostTime: postTime,
				})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkPosts(c)
		}
	}
	walkPosts(doc)

	if len(posts) > 0 {
		// First post is the problem statement.
		p.Content = posts[0].Content
		p.Author = posts[0].Author
		p.PostTime = posts[0].PostTime
		if len(posts) > 1 {
			p.Posts = posts[1:]
		}
	}
	p.Replies = len(posts)

	return p, nil
}

// ParseSearchResults parses a Discuz! search results page.
func ParseSearchResults(body []byte) ([]Thread, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var threads []Thread
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tbody" {
			if attr(n, "id") == "threadlisttableid" || attr(n, "id") == "search_result" {
				parseThreadListTable(n, "", &threads)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return threads, nil
}

// ParseForumIndex parses the main forum index and returns forum categories.
func ParseForumIndex(body []byte) ([]Forum, error) {
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	var forums []Forum
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if (n.Data == "td" || n.Data == "div") && strings.Contains(attr(n, "class"), "fl_inf") {
				f := parseForumBlock(n)
				if f.ID != "" {
					forums = append(forums, f)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return forums, nil
}

func parseForumBlock(n *html.Node) Forum {
	var f Forum
	findLinks(n, func(href, text string) bool {
		if idx := strings.Index(href, "fid="); idx >= 0 {
			part := href[idx+4:]
			if amp := strings.Index(part, "&"); amp >= 0 {
				part = part[:amp]
			}
			f.ID = part
			f.Name = strings.TrimSpace(text)
			f.URL = baseURL + "/forum.php?mod=forumdisplay&fid=" + f.ID
			return true
		}
		return false
	})
	return f
}

// FallbackForums returns a hardcoded list of known mathchina.com forum IDs
// used when the forum index cannot be parsed.
var FallbackForums = []Forum{
	{ID: "2", Name: "竞赛数学", URL: baseURL + "/forum.php?mod=forumdisplay&fid=2"},
	{ID: "7", Name: "竞赛题库", URL: baseURL + "/forum.php?mod=forumdisplay&fid=7"},
	{ID: "9", Name: "联赛试题", URL: baseURL + "/forum.php?mod=forumdisplay&fid=9"},
	{ID: "12", Name: "数学竞赛", URL: baseURL + "/forum.php?mod=forumdisplay&fid=12"},
	{ID: "15", Name: "CMO历年真题", URL: baseURL + "/forum.php?mod=forumdisplay&fid=15"},
	{ID: "20", Name: "希望杯", URL: baseURL + "/forum.php?mod=forumdisplay&fid=20"},
}

// --- date parsing ---

var dateRE = regexp.MustCompile(`(\d{4})-(\d{1,2})-(\d{1,2})(?:\s+(\d{1,2}):(\d{2}))?`)
var minutesAgoRE = regexp.MustCompile(`(\d+)\s*分钟前`)
var hoursAgoRE = regexp.MustCompile(`(\d+)\s*小时前`)

func parseDate(s string, now time.Time) (time.Time, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	s = strings.TrimSpace(s)

	// Full datetime: 2024-1-5 08:30
	if m := dateRE.FindStringSubmatch(s); m != nil {
		year, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		day, _ := strconv.Atoi(m[3])
		hour, min := 0, 0
		if m[4] != "" {
			hour, _ = strconv.Atoi(m[4])
			min, _ = strconv.Atoi(m[5])
		}
		return time.Date(year, time.Month(month), day, hour, min, 0, 0, loc), nil
	}
	// Yesterday
	if strings.Contains(s, "昨天") {
		rest := strings.TrimSpace(strings.TrimPrefix(s, "昨天"))
		yesterday := now.In(loc).AddDate(0, 0, -1)
		if rest != "" {
			if hm, err := time.ParseInLocation("15:04", rest, loc); err == nil {
				return time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(),
					hm.Hour(), hm.Minute(), 0, 0, loc), nil
			}
		}
		return yesterday.Truncate(24 * time.Hour), nil
	}
	// N minutes ago
	if m := minutesAgoRE.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		return now.Add(-time.Duration(n) * time.Minute), nil
	}
	// N hours ago
	if m := hoursAgoRE.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		return now.Add(-time.Duration(n) * time.Hour), nil
	}
	return time.Time{}, fmt.Errorf("unrecognized date: %q", s)
}

// --- HTML helpers ---

// attr returns the value of an HTML attribute.
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// findElement finds the first descendant element with the given tag name.
func findElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

// findByID finds the first element with the given ID attribute.
func findByID(n *html.Node, id string) *html.Node {
	if n.Type == html.ElementNode && attr(n, "id") == id {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findByID(c, id); found != nil {
			return found
		}
	}
	return nil
}

// findByClass finds the first element whose class contains the given string.
func findByClass(n *html.Node, class string) *html.Node {
	if n.Type == html.ElementNode && strings.Contains(attr(n, "class"), class) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findByClass(c, class); found != nil {
			return found
		}
	}
	return nil
}

// textContent extracts all text content from a node and its descendants.
func textContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}

// collectText collects all non-empty text segments from a node.
func collectText(n *html.Node) []string {
	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			s := strings.TrimSpace(n.Data)
			if s != "" {
				out = append(out, s)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return out
}

// findLinks calls fn for each <a> href and text in n's subtree.
// fn returns true to stop.
func findLinks(n *html.Node, fn func(href, text string) bool) {
	if n.Type == html.ElementNode && n.Data == "a" {
		href := attr(n, "href")
		text := textContent(n)
		if fn(href, text) {
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findLinks(c, fn)
	}
}

// cleanNum strips non-digit characters from a count string.
func cleanNum(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// cleanHTML converts an HTML node subtree to plain text with minimal formatting.
func cleanHTML(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "br", "p", "div":
				sb.WriteString("\n")
			case "blockquote":
				sb.WriteString("\n> ")
			}
			// Skip display:none elements (ads).
			style := attr(n, "style")
			if strings.Contains(strings.ToLower(style), "display:none") ||
				strings.Contains(strings.ToLower(style), "display: none") {
				return
			}
		}
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}
