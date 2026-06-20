package cli

import (
	"fmt"
	"os"

	"github.com/tamnd/mathchina-cli/mathchina"
	"github.com/tamnd/mathchina-cli/pkg/render"
)

// outputFlags holds the shared output format flags.
type outputFlags struct {
	format   string
	fields   []string
	noHeader bool
	tmpl     string
}

func (o *outputFlags) renderer(defaultFormat string) *render.Renderer {
	f := render.Format(o.format)
	if !f.Valid() {
		f = render.Format(defaultFormat)
	}
	return render.New(os.Stdout, f, o.fields, o.noHeader, o.tmpl)
}

// newMathchinaClient creates a mathchina client with default config.
func newMathchinaClient() *mathchina.Client {
	return mathchina.NewClient(mathchina.DefaultConfig())
}

// exitErr wraps an exit code.
type exitErr struct {
	code int
	msg  string
}

func (e *exitErr) Error() string { return e.msg }

func fatalf(code int, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, "cmo:", msg)
	return &exitErr{code: code, msg: msg}
}
