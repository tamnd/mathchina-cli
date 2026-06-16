package mathchina

import (
	"net/url"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// Domain is the mathchina kit domain, registered for binary identity and version output.
func init() { kit.Register(Domain{}) }

// Domain is the mathchina driver.
type Domain struct{}

// Info returns the domain identity.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "mathchina",
		Hosts:  []string{"www.mathchina.com", "mathchina.com"},
		Identity: kit.Identity{
			Binary: "cmo",
			Short:  "Browse Chinese math competition problems from mathchina.com",
			Long: `Browse Chinese math competition problems from mathchina.com (数学中国).

Fetch competition threads, problem details, and search the math forum.
No authentication required; all content is publicly accessible.`,
			Site: "www.mathchina.com",
			Repo: "https://github.com/tamnd/mathchina-cli",
		},
	}
}

// Register is a no-op; commands are added as kit escape-hatch commands in cli/root.go.
func (Domain) Register(app *kit.App) {}

// Classify turns a mathchina.com URL into a (type, id) pair.
func (Domain) Classify(input string) (uriType, id string, err error) {
	id = refPath(input)
	if id == "" {
		return "", "", errs.Usage("unrecognized mathchina reference: %q", input)
	}
	return "page", id, nil
}

// Locate returns the URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	if uriType != "page" {
		return "", errs.Usage("mathchina has no resource type %q", uriType)
	}
	return "http://www.mathchina.com/bbs/" + strings.Trim(id, "/"), nil
}

func refPath(input string) string {
	input = strings.TrimSpace(input)
	if u, err := url.Parse(input); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		return strings.Trim(u.Path, "/")
	}
	return strings.Trim(input, "/")
}
