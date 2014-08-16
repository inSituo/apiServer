package Api

import (
    "fmt"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/DBAccess"
    "github.com/inSituo/apiServer/LeveledLogger"
    "net/http"
)

type handler func(http.ResponseWriter, *http.Request)

type ApiSection struct {
    db     *DBAccess.DB
    log    *LeveledLogger.Logger
    r      *mux.Router
    prefix string
    base   string
}

func (as *ApiSection) setupRoute(method, endpoint string, f handler) {
    if as.r == nil {
        panic("'setupRoute' cannot be called before router is init'd.")
    }
    if as.log == nil {
        panic("'setupRoute' cannot be called before logger is init'd.")
    }
    route := genRoute(as.prefix, as.base, endpoint)
    as.log.Debugf("Setting up route GET %s", route)
    as.r.HandleFunc(route, f).Methods(method)
}

func genRoute(prefix, base, route string) string {
    return fmt.Sprintf("/%s%s/%s", prefix, base, route)
}
