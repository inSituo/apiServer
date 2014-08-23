package Middleware

import (
    "encoding/json"
    "encoding/xml"
    "github.com/gorilla/context"
    "github.com/inSituo/apiServer/LeveledLogger"
    "mime"
    "net/http"
    "strings"
)

const (
    _RESPONSE_CONTENT       = "__response_content"
    _RESPONSE_CODE          = "__response_code"
    RESPONSE_JSON           = iota
    RESPONSE_XML            = iota
    RESPONSE_DEFAULT_FORMAT = RESPONSE_JSON
)

type ResponseSetter func(req *http.Request, code int, content interface{})

type Responder struct {
    log    *LeveledLogger.Logger
    format int
}

// this middleware looks in the request context to find the response
// data to be written, and then writes it in one of two formats:
// 1. JSON
// 2. XML
func (r *Responder) Respond(res http.ResponseWriter, req *http.Request) {
    iname := "Responder.Respond"
    var marshalled []byte
    var ctype string
    var err error
    content := context.Get(req, _RESPONSE_CONTENT)
    switch r.format {
    case RESPONSE_JSON:
        if content != nil {
            marshalled, err = json.Marshal(content)
        }
        ctype = mime.TypeByExtension(".json")
    case RESPONSE_XML:
        if content != nil {
            marshalled, err = xml.Marshal(content)
        }
        ctype = mime.TypeByExtension(".xml")
    default:
        (&Responder{log: r.log, format: RESPONSE_DEFAULT_FORMAT}).Respond(res, req)
        return
    }
    if err == nil {
        code, found := context.Get(req, _RESPONSE_CODE).(int)
        if !found {
            r.log.Debug(iname, "response code not set. skipping auto response", context.Get(req, _RESPONSE_CODE))
            return
        }
        r.log.Debug(iname, "writing response", code)
        res.Header().Set("Content-Type", ctype)
        res.WriteHeader(code)
        res.Write(marshalled)
    } else {
        r.log.Warn(iname, "encoding error", err)
        res.WriteHeader(http.StatusInternalServerError)
    }
}

func SetResponse(req *http.Request, code int, content interface{}) {
    context.Set(req, _RESPONSE_CODE, code)
    context.Set(req, _RESPONSE_CONTENT, content)
}

func AutoFormatResponder(log *LeveledLogger.Logger, fmtHeaderName string) http.HandlerFunc {
    return func(res http.ResponseWriter, req *http.Request) {
        formatStr := strings.ToLower(req.Header.Get(fmtHeaderName))
        var format int
        switch formatStr {
        case "json":
            format = RESPONSE_JSON
        case "xml":
            format = RESPONSE_XML
        default:
            format = RESPONSE_DEFAULT_FORMAT
        }
        (&Responder{log: log, format: format}).Respond(res, req)
    }
}
