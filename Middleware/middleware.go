package Middleware

import (
    "encoding/xml"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/LeveledLogger"
    "gopkg.in/mgo.v2/bson"
    "net/http"
)

func IdVerifier(log *LeveledLogger.Logger, c *Chain, setRes ResponseSetter) http.HandlerFunc {
    iname := "Middleware.IdVerifier"
    return func(res http.ResponseWriter, req *http.Request) {
        id := mux.Vars(req)["id"]
        if id != "" {
            if !bson.IsObjectIdHex(id) {
                log.Debug(iname, "invalid id", id)
                log.Debug(iname, "breaking middleware chain")
                c.Break(req)
                setRes(req, http.StatusBadRequest, ErrRes{Reason: "invalid id"})
                return
            }
            log.Debug(iname, "valid id", id)
        }
    }
}

func RequestInfo(log *LeveledLogger.Logger, apiKeyHeader string) http.HandlerFunc {
    iname := "Middleware.RequestInfo"
    return func(res http.ResponseWriter, req *http.Request) {
        log.Info(iname, req.RemoteAddr, req.Header.Get(apiKeyHeader), req.Method, req.RequestURI)
    }
}

func RequestDebugInfo(log *LeveledLogger.Logger) http.HandlerFunc {
    iname := "Middleware.RequestDebugInfo"
    return func(res http.ResponseWriter, req *http.Request) {
        log.Debug(iname, "remote", req.RemoteAddr)
        log.Debug(iname, "method", req.Method)
        log.Debug(iname, "uri", req.RequestURI)
        for k, v := range req.Header {
            log.Debug(iname, "header", k, v)
        }
    }
}

type ErrRes struct {
    XMLName xml.Name `json:"-" xml:"error"`
    Reason  string   `json:"reason" xml:"reason"`
}
