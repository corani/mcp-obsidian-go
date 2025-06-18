package obsidian

import (
	"crypto/tls"
	"net/http"

	"github.com/corani/mcp-obsidian-go/internal/config"
)

type roundtripper struct {
	transport http.RoundTripper
	conf      *config.Config
}

func (r roundtripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+r.conf.ObsidianAPIKey)
	req.Header.Add("Accept", "application/vnd.olrapi.note+json")
	return r.transport.RoundTrip(req)
}

func newTransport(conf *config.Config) http.RoundTripper {
	return roundtripper{
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// TODO(daniel): allow option to provide the obsidian certificate.
				InsecureSkipVerify: true,
			},
		},
		conf: conf,
	}
}
