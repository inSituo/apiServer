package Api

import (
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "github.com/inSituo/apiServer/Middleware"
    "gopkg.in/mgo.v2/bson"
    "net/http"
)

type ApiSection struct {
    db  *DBAccess.DB
    log *LeveledLogger.Logger
    r   *mux.Router
    c   *Middleware.Chain
}

func (as *ApiSection) setupRoute(method, endpoint string, f http.HandlerFunc) {
    if as.r == nil {
        panic("'setupRoute' cannot be called before route is init'd.")
    }
    if as.log == nil {
        panic("'setupRoute' cannot be called before logger is init'd.")
    }
    as.log.Debugf("Setting up route %s %s", method, endpoint)
    g := f
    if as.c != nil {
        as.c.Push(f)
        g = as.c.MakeHandler()
        as.c.Pop()
    }
    as.r.HandleFunc(endpoint, g).Methods(method)
}

func (as *ApiSection) use(f http.HandlerFunc) {
    if as.c == nil {
        as.c = &Middleware.Chain{}
    }
    as.c.Push(f)
}

type ContentRevision struct {
    uid     bson.ObjectId `bson:"uid"`
    ts      int           `bson:"ts"`
    lat     float32       `bson:"lat"`
    lon     float32       `bson:"lon"`
    address string        `bson:"address"`
    content string        `bson:"content"`
}
