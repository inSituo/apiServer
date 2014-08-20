package Middleware

import (
    "encoding/json"
    "github.com/gorilla/context"
    "github.com/gorilla/mux"
    "github.com/inSituo/apiServer/LeveledLogger"
    "gopkg.in/mgo.v2/bson"
    "net/http"
)

func IdVerifier(log *LeveledLogger.Logger) http.HandlerFunc {
    return func(res http.ResponseWriter, req *http.Request) {
        id := mux.Vars(req)["id"]
        if id != "" {
            if !bson.IsObjectIdHex(id) {
                log.Debugf("ID %s is invalid. Breaking middleware chain", id)
                context.Set(req, "break-chain", true)
                res.WriteHeader(http.StatusBadRequest)
                js, _ := json.Marshal(map[string]string{
                    "error": "invalid id",
                })
                res.Write(js)
                return
            }
            log.Debugf("ID %s is valid.", id)
        }
    }
}

func RequestDebugInfo(log *LeveledLogger.Logger) http.HandlerFunc {
    return func(res http.ResponseWriter, req *http.Request) {
        log.Debugf("--- New HTTP request ---")
        log.Debugf("Remote: %s", req.RemoteAddr)
        log.Debugf("Method: %s", req.Method)
        log.Debugf("URI: %s", req.RequestURI)
        for k, v := range req.Header {
            log.Debugf("Header %s: %s", k, v)
        }
        log.Debugf("------------------------")
    }
}
