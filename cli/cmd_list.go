package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/mathchina-cli/mathchina"
)

type listCmd struct {
	out         outputFlags
	fid         string
	year        int
	competition string
	page        int
	pages       int
	limit       int
}

func newListCmd() kit.Command {
	c := &listCmd{}
	return kit.Command{
		Use:   "list",
		Short: "List competition threads from the math forum",
		Long: `List competition problem threads from mathchina.com.

Fetches threads from a forum section and optionally filters by year or
competition name. Supports multi-page fetching with --pages.

Examples:
  cmo list
  cmo list --fid 15 --year 2023
  cmo list --competition CMO --pages 3
  cmo list --year 2024 -o jsonl | jq .title`,
		Args:  kit.NoArgs,
		Flags: c.flags,
		Run:   c.run,
	}
}

func (c *listCmd) flags(f *kit.FlagSet) {
	f.StringVarP(&c.out.format, "output", "o", "table", "Output format: table|json|jsonl|csv|tsv|raw")
	f.StringSliceVar(&c.out.fields, "fields", nil, "Select and reorder output columns")
	f.BoolVar(&c.out.noHeader, "no-header", false, "Suppress header row")
	f.StringVar(&c.out.tmpl, "template", "", "Go text/template applied per record")
	f.IntVar(&c.limit, "limit", 0, "Stop after N results (0 = all)")
	f.StringVar(&c.fid, "fid", "7", "Forum ID to list from")
	f.IntVar(&c.year, "year", 0, "Filter threads containing this year in their title")
	f.StringVar(&c.competition, "competition", "", "Filter threads containing this competition name (case-insensitive)")
	f.IntVarP(&c.page, "page", "p", 1, "Start page number")
	f.IntVar(&c.pages, "pages", 1, "Number of pages to fetch")
}

func (c *listCmd) run(ctx context.Context, _ []string) error {
	cfg := mathchina.DefaultConfig()
	client := mathchina.NewClient(cfg)

	var all []mathchina.Thread
	for pg := c.page; pg < c.page+c.pages; pg++ {
		url := fmt.Sprintf("%s/forum.php?mod=forumdisplay&fid=%s&page=%d", cfg.BaseURL, c.fid, pg)
		body, err := client.GetPage(ctx, url)
		if err != nil {
			return fatalf(5, "fetch page %d: %v", pg, err)
		}
		threads, err := mathchina.ParseThreadList(body, c.fid)
		if err != nil {
			return fatalf(1, "parse page %d: %v", pg, err)
		}
		all = append(all, threads...)
	}

	// Apply filters.
	yearStr := ""
	if c.year > 0 {
		yearStr = strconv.Itoa(c.year)
	}
	rend := c.out.renderer("table")
	count := 0
	for _, th := range all {
		if yearStr != "" && !strings.Contains(th.Title, yearStr) {
			continue
		}
		if c.competition != "" && !strings.Contains(strings.ToLower(th.Title), strings.ToLower(c.competition)) {
			continue
		}
		if err := rend.Render(th.ToRecord()); err != nil {
			return err
		}
		count++
		if c.limit > 0 && count >= c.limit {
			break
		}
	}
	if count == 0 {
		return fatalf(3, "no threads found")
	}
	return nil
}
