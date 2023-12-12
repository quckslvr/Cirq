package cirq

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/julienschmidt/httprouter"
)

type ErrorHandler func(error, *Context) error

type Context struct {
	response http.ResponseWriter
	request  *http.Request
	ctx      context.Context
}

func (c *Context) Render(component templ.Component) error {
	return component.Render(c.ctx, c.response)
}

type Tide func(Handler) Handler

type Handler func(c *Context) error

type Cirq struct {
	ErrorHandler ErrorHandler
	router       *httprouter.Router
	middlewares  []Tide
}

func New() *Cirq {
	return &Cirq{
		router:       httprouter.New(),
		ErrorHandler: defaultErrorHandler,
	}
}

func (c *Cirq) Tide(tides ...Tide) {
	c.middlewares = append(c.middlewares, tides...)
}

func (c *Cirq) Start(addr string) error {
	return http.ListenAndServe(addr, c.router)
}

func (c *Cirq) Get(path string, h Handler, tides ...Tide) {
	c.router.GET(path, c.makeHTTPRouterHandler(h))
}

func (c *Cirq) makeHTTPRouterHandler(h Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		ctx := &Context{
			response: w,
			request:  r,
			ctx:      context.Background(),
		}
		if err := h(ctx); err != nil {
			//todo handle the error feom the error handler?
			c.ErrorHandler(err, ctx)
		}
	}
}

func defaultErrorHandler(err error, c *Context) error {
	slog.Error("error", err)
	return nil
}
