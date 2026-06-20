package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/mathchina-cli/mathchina"
)

type problemCmd struct {
	out      outputFlags
	allPages bool
}

func newProblemCmd() kit.Command {
	c := &problemCmd{}
	return kit.Command{
		Use:   "problem <id>",
		Short: "Fetch a problem thread by its ID",
		Long: `Fetch the content of a specific math problem thread.

The ID is the numeric thread ID from the URL (tid=N).
By default, only the first page is fetched. Use --all-pages for full threads.

Examples:
  cmo problem 12345
  cmo problem 12345 --all-pages -o json
  cmo problem 12345 | jq .content`,
		Args:  kit.ExactArgs(1),
		Flags: c.flags,
		Run:   c.run,
	}
}

func (c *problemCmd) flags(f *kit.FlagSet) {
	f.StringVarP(&c.out.format, "output", "o", "json", "Output format: table|json|jsonl|csv|tsv|raw")
	f.StringSliceVar(&c.out.fields, "fields", nil, "Select and reorder output columns")
	f.BoolVar(&c.out.noHeader, "no-header", false, "Suppress header row")
	f.StringVar(&c.out.tmpl, "template", "", "Go text/template applied per record")
	f.BoolVar(&c.allPages, "all-pages", false, "Fetch all reply pages")
}

func (c *problemCmd) run(ctx context.Context, args []string) error {
	tid := args[0]
	cfg := mathchina.DefaultConfig()
	client := mathchina.NewClient(cfg)

	url := fmt.Sprintf("%s/forum.php?mod=viewthread&tid=%s", cfg.BaseURL, tid)
	body, err := client.GetPage(ctx, url)
	if err != nil {
		return fatalf(5, "fetch thread %s: %v", tid, err)
	}

	p, err := mathchina.ParseThreadDetail(body, tid)
	if err != nil {
		return fatalf(1, "parse thread %s: %v", tid, err)
	}
	if p.Title == "" && p.Content == "" {
		return fatalf(3, "thread %s not found", tid)
	}

	rec := p.ToRecord()
	format := c.out.format
	if format == "" {
		format = "json"
	}
	if format == "json" {
		data, err := json.MarshalIndent(rec, "", "  ")
		if err != nil {
			return fatalf(1, "encode: %v", err)
		}
		_, _ = fmt.Fprintln(os.Stdout, string(data))
		return nil
	}
	return c.out.renderer("json").Render(rec)
}
