package cli

import (
	"context"
	"fmt"
	"net/url"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/mathchina-cli/mathchina"
)

type searchCmd struct {
	out   outputFlags
	fid   string
	page  int
	pages int
	limit int
}

func newSearchCmd() kit.Command {
	c := &searchCmd{}
	return kit.Command{
		Use:   "search <query>",
		Short: "Search the forum for math problems",
		Long: `Search mathchina.com for threads matching a query.

Searches the Discuz! forum search and returns matching threads.
Chinese and English queries are both supported.

Examples:
  cmo search "CMO 2023"
  cmo search "几何" --fid 7
  cmo search "希望杯" --pages 3 -o jsonl`,
		Args:  kit.ExactArgs(1),
		Flags: c.flags,
		Run:   c.run,
	}
}

func (c *searchCmd) flags(f *kit.FlagSet) {
	f.StringVarP(&c.out.format, "output", "o", "table", "Output format: table|json|jsonl|csv|tsv|raw")
	f.StringSliceVar(&c.out.fields, "fields", nil, "Select and reorder output columns")
	f.BoolVar(&c.out.noHeader, "no-header", false, "Suppress header row")
	f.StringVar(&c.out.tmpl, "template", "", "Go text/template applied per record")
	f.IntVar(&c.limit, "limit", 0, "Max results (0 = all)")
	f.StringVar(&c.fid, "fid", "", "Restrict search to forum ID (optional)")
	f.IntVarP(&c.page, "page", "p", 1, "Start result page number")
	f.IntVar(&c.pages, "pages", 1, "Number of search result pages to fetch")
}

func (c *searchCmd) run(ctx context.Context, args []string) error {
	query := args[0]
	cfg := mathchina.DefaultConfig()
	client := mathchina.NewClient(cfg)

	var all []mathchina.Thread
	for pg := c.page; pg < c.page+c.pages; pg++ {
		searchURL := fmt.Sprintf(
			"%s/search.php?mod=forum&orderby=lastpost&ascdesc=desc&searchsubmit=yes&kw=%s&page=%d",
			cfg.BaseURL,
			url.QueryEscape(query),
			pg,
		)
		if c.fid != "" {
			searchURL += "&fid=" + c.fid
		}
		body, err := client.GetPage(ctx, searchURL)
		if err != nil {
			return fatalf(5, "fetch search page %d: %v", pg, err)
		}
		threads, err := mathchina.ParseSearchResults(body)
		if err != nil {
			return fatalf(1, "parse search page %d: %v", pg, err)
		}
		all = append(all, threads...)
	}

	rend := c.out.renderer("table")
	count := 0
	for _, th := range all {
		if err := rend.Render(th.ToRecord()); err != nil {
			return err
		}
		count++
		if c.limit > 0 && count >= c.limit {
			break
		}
	}
	return nil
}
