package mathchina

import "time"

// Thread represents a forum thread in a listing.
type Thread struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	Replies      int       `json:"replies"`
	Views        int       `json:"views"`
	LastPostTime time.Time `json:"last_post_time"`
	ForumID      string    `json:"forum_id"`
	URL          string    `json:"url"`
}

// Post represents one reply in a thread.
type Post struct {
	PostID   string    `json:"post_id"`
	Author   string    `json:"author"`
	Content  string    `json:"content"`
	PostTime time.Time `json:"post_time"`
}

// Problem is a thread with its full content and posts.
type Problem struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Author   string    `json:"author"`
	PostTime time.Time `json:"post_time"`
	Content  string    `json:"content"`
	Replies  int       `json:"replies"`
	Views    int       `json:"views"`
	URL      string    `json:"url"`
	Posts    []Post    `json:"posts"`
}

// Forum is a category in the forum index.
type Forum struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Threads     int    `json:"threads"`
	Posts       int    `json:"posts"`
	URL         string `json:"url"`
}

// Output record types (formatted for pkg/render).

// ThreadRecord is one row in list/search output.
type ThreadRecord struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Author       string `json:"author"`
	Replies      int    `json:"replies"`
	Views        int    `json:"views"`
	LastPostTime string `json:"last_post_time"`
	ForumID      string `json:"forum_id"`
	URL          string `json:"url"`
}

// PostRecord is one reply in a problem output.
type PostRecord struct {
	PostID   string `json:"post_id"`
	Author   string `json:"author"`
	Content  string `json:"content"`
	PostTime string `json:"post_time"`
}

// ProblemRecord is the output of the `problem` command.
type ProblemRecord struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	Author   string       `json:"author"`
	PostTime string       `json:"post_time"`
	Content  string       `json:"content"`
	Replies  int          `json:"replies"`
	Views    int          `json:"views"`
	URL      string       `json:"url"`
	Posts    []PostRecord `json:"posts"`
}

// ForumRecord is one row in the `forums` output.
type ForumRecord struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Threads     int    `json:"threads"`
	Posts       int    `json:"posts"`
	URL         string `json:"url"`
}

// ToRecord converts a Thread to a ThreadRecord.
func (t Thread) ToRecord() ThreadRecord {
	ts := ""
	if !t.LastPostTime.IsZero() {
		ts = t.LastPostTime.UTC().Format("2006-01-02T15:04:05Z")
	}
	return ThreadRecord{
		ID:           t.ID,
		Title:        t.Title,
		Author:       t.Author,
		Replies:      t.Replies,
		Views:        t.Views,
		LastPostTime: ts,
		ForumID:      t.ForumID,
		URL:          t.URL,
	}
}

// ToRecord converts a Problem to a ProblemRecord.
func (p Problem) ToRecord() ProblemRecord {
	ts := ""
	if !p.PostTime.IsZero() {
		ts = p.PostTime.UTC().Format("2006-01-02T15:04:05Z")
	}
	posts := make([]PostRecord, len(p.Posts))
	for i, post := range p.Posts {
		pts := ""
		if !post.PostTime.IsZero() {
			pts = post.PostTime.UTC().Format("2006-01-02T15:04:05Z")
		}
		posts[i] = PostRecord{
			PostID:   post.PostID,
			Author:   post.Author,
			Content:  post.Content,
			PostTime: pts,
		}
	}
	return ProblemRecord{
		ID:       p.ID,
		Title:    p.Title,
		Author:   p.Author,
		PostTime: ts,
		Content:  p.Content,
		Replies:  p.Replies,
		Views:    p.Views,
		URL:      p.URL,
		Posts:    posts,
	}
}
