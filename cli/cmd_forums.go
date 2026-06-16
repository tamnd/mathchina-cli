package cli

import (
	"context"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/mathchina-cli/mathchina"
)

type forumsCmd struct {
	out outputFlags
}

func newForumsCmd() kit.Command {
	c := &forumsCmd{}
	return kit.Command{
		Use:   "forums",
		Short: "List the competition math forums",
		Long: `List all competition math forum sections on mathchina.com.

Fetches the main forum index and lists forum sections with their IDs
and descriptions. Falls back to a known list when the forum index is
unavailable.

Examples:
  cmo forums
  cmo forums -o json
  cmo forums -o jsonl | jq .id`,
		Args:  kit.NoArgs,
		Flags: c.flags,
		Run:   c.run,
	}
}

func (c *forumsCmd) flags(f *kit.FlagSet) {
	f.StringVarP(&c.out.format, "output", "o", "table", "Output format: table|json|jsonl|csv|tsv|raw")
	f.StringSliceVar(&c.out.fields, "fields", nil, "Select and reorder output columns")
	f.BoolVar(&c.out.noHeader, "no-header", false, "Suppress header row")
	f.StringVar(&c.out.tmpl, "template", "", "Go text/template applied per record")
}

func (c *forumsCmd) run(ctx context.Context, _ []string) error {
	cfg := mathchina.DefaultConfig()
	client := mathchina.NewClient(cfg)

	body, err := client.GetPage(ctx, cfg.BaseURL+"/")
	forums := mathchina.FallbackForums
	if err == nil {
		parsed, perr := mathchina.ParseForumIndex(body)
		if perr == nil && len(parsed) > 0 {
			forums = parsed
		}
	}

	rend := c.out.renderer("table")
	for _, f := range forums {
		rec := mathchina.ForumRecord{
			ID:          f.ID,
			Name:        f.Name,
			Description: f.Description,
			Threads:     f.Threads,
			Posts:       f.Posts,
			URL:         f.URL,
		}
		if err := rend.Render(rec); err != nil {
			return err
		}
	}
	return nil
}
