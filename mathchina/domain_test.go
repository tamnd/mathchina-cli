package mathchina

import (
	"testing"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "mathchina" {
		t.Errorf("Scheme = %q, want mathchina", info.Scheme)
	}
	if info.Identity.Binary != "cmo" {
		t.Errorf("Identity.Binary = %q, want cmo", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"forum.php?fid=7", "page", "forum.php?fid=7"},
		{"http://www.mathchina.com/bbs/forum.php", "page", "bbs/forum.php"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("page", "forum.php?fid=7")
	if err != nil {
		t.Fatalf("Locate: %v", err)
	}
	if got != "http://www.mathchina.com/bbs/forum.php?fid=7" {
		t.Errorf("Locate = %q", got)
	}
}
