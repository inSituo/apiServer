package MiddlewareChain

import (
    "net/http"
)

type httpHandler func(http.ResponseWriter, *http.Request)

type middleware struct {
    handler httpHandler
    next    *middleware
}

type Chain struct {
    mws  []*middleware
    last *middleware
}

func (c *Chain) pushMiddleware(m *middleware) {
    if c.last != nil {
        c.last.next = m
        c.mws = append(c.mws, c.last)
    }
    c.last = m
}

func (c *Chain) Push(f httpHandler) {
    m := &middleware{
        handler: f,
        next:    nil,
    }
    c.pushMiddleware(m)
}

func (c *Chain) PushHandler(h http.Handler) {
    c.Push(h.ServeHTTP)
}

func (c *Chain) MakeHandler() http.HandlerFunc {
    if c.last == nil {
        panic("Must have at least one middleware in the chain to make a handler")
    }
    h := http.HandlerFunc(c.last.handler)
    for i := len(c.mws) - 1; i >= 0; i-- {
        h = func(parent, child httpHandler) http.HandlerFunc {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                parent(w, r)
                child(w, r)
            })
        }(c.mws[i].handler, h.ServeHTTP)
    }
    return h
}
