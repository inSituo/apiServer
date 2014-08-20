package Middleware

import (
    "github.com/gorilla/context"
    "net/http"
)

type middleware struct {
    handler http.HandlerFunc
    next    *middleware
}

type Chain struct {
    mws       []*middleware
    last      *middleware
    breakable bool
}

func NewChain(breakable bool) *Chain {
    return &Chain{breakable: breakable}
}

func (c *Chain) pushMiddleware(m *middleware) *Chain {
    if c.last != nil {
        c.last.next = m
        c.mws = append(c.mws, c.last)
    }
    c.last = m
    return c
}

func (c *Chain) Push(f http.HandlerFunc) *Chain {
    m := &middleware{
        handler: f,
        next:    nil,
    }
    return c.pushMiddleware(m)
}

func (c *Chain) Pop() (f http.HandlerFunc) {
    f = c.last.handler
    c.last = c.mws[len(c.mws)-1]
    c.last.next = nil
    c.mws = c.mws[:len(c.mws)-1]
    return
}

func (c *Chain) PushHandler(h http.Handler) *Chain {
    return c.Push(h.ServeHTTP)
}

func (c *Chain) MakeHandler() http.HandlerFunc {
    if c.last == nil {
        panic("Must have at least one middleware in the chain to make a handler")
    }
    h := http.HandlerFunc(c.last.handler)
    for i := len(c.mws) - 1; i >= 0; i-- {
        h = func(parent, child http.HandlerFunc) http.HandlerFunc {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                parent(w, r)
                if !c.breakable || context.Get(r, "break-chain") != true {
                    child(w, r)
                }
                context.Delete(r, "break-chain")
            })
        }(c.mws[i].handler, h.ServeHTTP)
    }
    return h
}
