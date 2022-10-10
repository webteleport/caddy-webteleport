package teleport

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/sirupsen/logrus"
	"github.com/webteleport/ufo"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(Middleware{})
	httpcaddyfile.RegisterHandlerDirective("teleport", parseCaddyfile)
	logrus.Info("webteleport module inited")
}

type Middleware struct {
	logger  *zap.Logger
	Station string `json:"duration,omitempty"`
}

func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.http_webteleport",
		New: func() caddy.Module { return new(Middleware) },
	}
}

func (m *Middleware) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger(m)
	m.Station = "https://ufo.k0s.io"
	m.logger.Info("webteleport module inited")
	return nil
}

func (m *Middleware) Validate() error {
	return nil
}

type chandler struct {
	caddyhttp.Handler
}

func (c *chandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.Handler.ServeHTTP(w, r)
}

func IntoHandler(c caddyhttp.Handler) http.Handler {
	return &chandler{c}
}

func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m.logger.Info("http_webteleport start")
	defer m.logger.Info("http_webteleport end")
	return ufo.Serve(m.Station, IntoHandler(next))
}

func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() && d.NextArg() {
		m.Station = d.Val()
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Middleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Middleware)(nil)
	_ caddy.Validator             = (*Middleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*Middleware)(nil)
	_ caddyfile.Unmarshaler       = (*Middleware)(nil)
)
