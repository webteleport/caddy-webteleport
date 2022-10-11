package teleport

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	logger     *zap.Logger
	Station    string `json:"duration,omitempty"`
	KnockTimes string `json:"knock_time,omitempty"`
	KnockURL   string `json:"knock_url,omitempty"`
}

func attachReplacerContext(r *http.Request) *http.Request {
	newCtx := context.WithValue(context.Background(), caddy.ReplacerCtxKey, caddy.NewEmptyReplacer())
	newCtx = context.WithValue(newCtx, caddyhttp.ServerCtxKey, Server)
	newCtx = context.WithValue(newCtx, "route_group", RouteGroup)
	newCtx = context.WithValue(newCtx, caddyhttp.OriginalRequestCtxKey, OriginalRequest)
	return r.WithContext(newCtx)
}

func (m *Middleware) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := attachReplacerContext(r)
		_ = r2.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
		_ = r2.Context().Value(caddyhttp.ServerCtxKey).(*caddyhttp.Server)
		_ = r2.Context().Value(caddyhttp.OriginalRequestCtxKey).(http.Request)
		_ = r2.Context().Value("route_group").(map[string]struct{})
		if Next == nil {
			m.logger.Info("not found")
			http.NotFoundHandler().ServeHTTP(w, r2)
			return
		}
		Next.ServeHTTP(w, r2)
	})
}

func (Middleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.http_webteleport",
		New: func() caddy.Module { return new(Middleware) },
	}
}

func (m *Middleware) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger(m)
	m.logger.Info("webteleport module inited: " + m.Station)
	go ufo.Serve(m.Station, m.Handler())
	return nil
}

func (m *Middleware) Validate() error {
	return nil
}

var Next caddyhttp.Handler
var Server *caddyhttp.Server
var RouteGroup map[string]struct{}
var Ctx context.Context
var OriginalRequest http.Request

func (m Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m.logger.Info("http_webteleport start")
	defer m.logger.Info("http_webteleport end")
	if Next != nil {
		m.logger.Info("get")
		return Next.ServeHTTP(w, r)
	} else {
		m.logger.Info("set")
		Next = next
		Server = r.Context().Value(caddyhttp.ServerCtxKey).(*caddyhttp.Server)
		RouteGroup = r.Context().Value("route_group").(map[string]struct{})
		OriginalRequest = r.Context().Value(caddyhttp.OriginalRequestCtxKey).(http.Request)
		for k, _ := range RouteGroup {
			println(k)
		}
		Ctx = r.Context()
		return Next.ServeHTTP(w, r)
	}
}

func httpGet(u string, t string) {
	n, err := strconv.Atoi(t)
	if err != nil {
		n = 1
	}
	for c := 0; c < n; c++ {
		// println(c)
		time.Sleep(time.Second)
		http.Get(u)
	}
}

func (m *Middleware) addDirective(directive string, arg string, pos int) {
	// m.logger.Info("http_webteleport add directive: " + directive + arg)
	if directive == "knock" {
		switch pos {
		case 1:
			m.KnockURL = arg
			m.KnockTimes = "1"
		case 2:
			m.KnockTimes = arg
		}
	}
	// m.logger.Info("http_webteleport unknown directive: " + directive)
}

func (m *Middleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() && d.NextArg() {
		m.Station = d.Val()
	}
	var directive string = ""
	_ = directive
	for i := 0; d.Next(); i++ {
		val := d.Val()
		if val == "knock" {
			i = 0
		}
		fmt.Sprintf("using %d %s", i, val)
		switch i {
		case 0:
			directive = val
		case 1, 2:
			m.addDirective(directive, val, i)
		}
	}
	go httpGet(m.KnockURL, m.KnockTimes)
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
